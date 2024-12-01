package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/momokii/ss-watcher/internal/database"
	"github.com/momokii/ss-watcher/internal/models"
	"github.com/momokii/ss-watcher/internal/repository"
	"github.com/momokii/ss-watcher/pkg/gdrive"
	"github.com/momokii/ss-watcher/pkg/utils"
)

func main() {
	var PATH, USER_EMAIL string

	// input path from user
	fmt.Println("Enter the absolute path to the screenshot folder you want to watch (ex: C:/Users/ACER): ")
	fmt.Scanln(&PATH)
	if PATH == "" {
		fmt.Println("The path is empty")
		return
	}

	// check and convert absolute path to forward slash and also check if the path exist or not on local machine
	PATH = filepath.ToSlash(PATH)

	absPath, err := filepath.Abs(PATH)
	if err != nil {
		fmt.Println("Error Abs Path: ", err)
		return
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("The inputted path ('%s') does not exist, try again!", absPath)
		return
	}

	// input email user to give permission to the folder on gdrive
	fmt.Println("\nEnter the email of the user you want to give permission to access the folder on GDrive: ")
	fmt.Scanln(&USER_EMAIL)
	if USER_EMAIL == "" {
		fmt.Println("The email is empty")
		return
	}
	// simple email checker structure
	valid, _ := utils.IsEmailFormatValid(USER_EMAIL)
	if !valid {
		fmt.Println("Invalid email format")
		return
	}

	// * ------------ WATCHER PROCESS INIT
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	if err = watcher.Add(PATH); err != nil {
		panic(err)
	}

	fmt.Println("\nWatching: " + PATH + " \n")

	// * ------------ GDRIVE PROCESS INIT
	gdrive := gdrive.NewGDrive("YOUR SERVICE ACCOUNT JSON PATH HERE...") // init process
	gdriveService := gdrive.GetService()

	// * ------------ INIT DATABASE PROCESS INIT
	db := database.InitDB()
	recordRepo := repository.NewRecordsRepository()
	permissionRepo := repository.NewUserPermission()
	fmt.Println("\n")

	// * ------------ GDRIVE PROCESS CHECKER FOLDER AND PERMISSION ACCESS
	// check base folder on gdrive exist or not
	// if not exist create base folder for upload the ss file
	var BASE_GRDRIVE_FOLDER_ID string
	BASE_GDRIVE_NAME := "SS-Watcher-Backup-GDrive-Folder"

	// check BASE FOLDER exist or not
	id, err := gdrive.CheckFolderExist(BASE_GDRIVE_NAME, "")
	if err != nil {
		fmt.Println("Error Check Folder Exist: ", err)
		return
	}

	// start tx for permission access process
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Error Begin Transaction: ", err)
		return
	}

	// if BASE_FOLDER not exist, create folder on root gdrive
	if id == "" {
		fmt.Println("Base Folder not exist, creating base folder...")

		// create base folder
		BASE_GRDRIVE_FOLDER_ID, err = gdrive.CreateFolder(BASE_GDRIVE_NAME, "")
		if err != nil {
			fmt.Println("Error Create Base Folder: ", err)
			return
		}

		// automatically add permission to the folder to the user email inputted
		// so the owner can access the folder on their gdrive
		if _, err := gdrive.NewUserPermission(BASE_GRDRIVE_FOLDER_ID, USER_EMAIL); err != nil {
			fmt.Println("Error Create Permission: ", err)
			return
		} else {
			// also add the user email to the db user permission
			permission_id, err := gdrive.NewUserPermission(BASE_GRDRIVE_FOLDER_ID, USER_EMAIL)
			if err != nil {
				fmt.Println("Error Create Permission: ", err)
			} else {
				// add new data to db user permission
				if err := permissionRepo.Create(tx, &models.UserPermission{
					PermissionID: permission_id,
					Email:        USER_EMAIL,
				}); err != nil {
					fmt.Println("Error Create Permission: ", err)
				}
			}
		}

		fmt.Println("Base Folder Created and Permission added for User: ", USER_EMAIL)

	} else {
		BASE_GRDRIVE_FOLDER_ID = id
		fmt.Println("Base Folder Exist")

		// check all permission on the folder
		permissions, err := gdriveService.Permissions.List(BASE_GRDRIVE_FOLDER_ID).SupportsAllDrives(true).Do()
		if err != nil {
			fmt.Println("Error listing permissions: ", err)
			return
		}

		all_user := make([]string, 0) // slice to store all user id permission

		// loop through all permission and check if the role is user and if user add id to slice
		for _, perm := range permissions.Permissions {
			// fmt.Printf("Permission ID: %s, Role: %s, Type: %s\n", perm.Id, perm.Role, perm.Type)

			if perm.Role == "writer" {
				// use single quote for each id to use 'IN' query on sql
				all_user = append(all_user, `'`+string(perm.Id)+`'`)
			}
		}

		// if slice > 0, so there is user permission on the folder
		if len(all_user) > 0 {
			// check email user permission on db
			permissions, err := permissionRepo.FindByID(tx, all_user)
			if err != nil {
				fmt.Println("Error Find By ID: ", err)
			} else {
				// if not found data from all user permission id from gdrive, add to db email inputted before and give permission to the folder on gdrive

				// check if USER EMAIL is already registered on the permission
				is_granted := false
				for _, perm := range *permissions { // loop through all permission on the folder
					if perm.Email == USER_EMAIL {
						is_granted = true
						break
					}
				}

				// add permission on gdrive folder for user if the USER EMAIL not found on gdrive permission list
				if !is_granted {
					fmt.Printf("User '%s' not found on GDrive Permission, adding permission...\n", USER_EMAIL)
					permission_id, err := gdrive.NewUserPermission(BASE_GRDRIVE_FOLDER_ID, USER_EMAIL)
					if err != nil {
						fmt.Println("Error Create Permission: ", err)
					} else {
						// add new data to db user permission
						if err := permissionRepo.Create(tx, &models.UserPermission{
							PermissionID: permission_id,
							Email:        USER_EMAIL,
						}); err != nil {
							fmt.Println("Error Create Permission: ", err)
						}
					}
				}

				fmt.Printf("User '%s' have permission to the folder on GDrive\n", USER_EMAIL)
			}

		} else {
			// if not found data on gdrive permission, add to db email inputted before and give permission to the folder on gdrive
			permission_id, err := gdrive.NewUserPermission(BASE_GRDRIVE_FOLDER_ID, USER_EMAIL)
			if err != nil {
				fmt.Println("Error Create Permission: ", err)
			} else {
				// add new data to db user permission
				err := permissionRepo.Create(tx, &models.UserPermission{
					PermissionID: permission_id,
					Email:        USER_EMAIL,
				})

				if err != nil {
					fmt.Println("Error Create Permission: ", err)
				} else {
					fmt.Printf("User '%s' added to the base folder\n", USER_EMAIL)
				}
			}
		}

	}

	// check to commit tx above and will return just if error appear on the process
	if p := recover(); p != nil {
		tx.Rollback()
		return
	} else if err != nil {
		if rbErr := tx.Rollback; rbErr != nil {
			fmt.Println("Error Rollback: ", rbErr)
		}
		return
	} else {
		if cErr := tx.Commit(); cErr != nil {
			fmt.Println("Error Commit: ", cErr)
		}
	}

	// * ------------ WATCHER PROCESS MAIN LOOP
	fmt.Println("\nWaiting for event...")
	for {
		select {
		case event := <-watcher.Events:
			fmt.Println("Event: ", event)

			filepath := event.Name

			filename := filepath[len(PATH)+1:] // +1 to remove the slash

			// ! --- WATCHER UPLOAD/NEW EVENT FILE PROCESS
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("Modified file: ", filepath)

				file, err := os.Open(filepath)
				if err != nil {
					fmt.Println("Error Open File: ", err)
					continue
				}

				// if success, add also file to gdrive folder
				// upload file
				folderId, err := gdrive.CheckExistOrCreateFolderSSDaily(BASE_GRDRIVE_FOLDER_ID)
				if err != nil {
					fmt.Println("Error Check Exist or Create Folder: ", err)
					continue
				} else {
					// upload file to gdrive folder
					fileUpload, err := gdrive.UploadFileDrive(filename, filepath, "image/png", folderId)
					if err != nil {
						fmt.Println("Error Upload File Drive: ", err)
					} else {

						fmt.Println("Upload File Success ID: ", fileUpload.Id)

						tx, err := db.Begin()
						if err != nil {
							fmt.Println("Error Begin Transaction: ", err)
							continue
						}

						dataFile := models.Records{
							ItemID:   fileUpload.Id,
							Name:     filename,
							FolderID: folderId,
							Date:     time.Now().String(),
						}

						if err := recordRepo.Create(tx, &dataFile); err != nil {
							fmt.Println("Error Create Record: ", err)
						}

						// because on loop, need to commit tx on manual wwithout defer
						if p := recover(); p != nil {
							tx.Rollback()
						} else if err != nil {
							if rbErr := tx.Rollback; rbErr != nil {
								fmt.Println("Error Rollback: ", rbErr)
							}
						} else {
							cErr := tx.Commit()
							if cErr != nil {
								fmt.Println("Error Commit: ", cErr)
							} else {
								fmt.Println("Store Record Success ID File: ", fileUpload.Id)
							}
						}
					}
				}

				file.Close()

				// ! --- WATCHER DELETE EVENT FILE PROCESS
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				fmt.Println("Remove file: ", filepath)

				tx, err := db.Begin()
				if err != nil {
					fmt.Println("Error Begin Transaction: ", err)
					continue
				}

				// first get file id from db based on filename
				itemData, err := recordRepo.FindByName(tx, filename)
				if err != nil && err != sql.ErrNoRows {
					fmt.Println("Error Find By Name: ", err)
					continue
				}

				if err == sql.ErrNoRows {
					fmt.Println("Data not found on DB")
					continue
				} else {
					// if exist, delete file from gdrive
					err = gdrive.DeleteFileDrive(itemData.ItemID)
					if err != nil {
						fmt.Println("Error Delete File Drive: ", err)
						continue
					} else {

						fmt.Println("Delete File from Drive Success ID: ", itemData.ItemID)

						// success delete from drive, continue delete data from db
						if err := recordRepo.Delete(tx, itemData.ItemID); err != nil {
							fmt.Println("Error Delete Record: ", err)
						}
					}
				}

				// commit tx process
				if p := recover(); p != nil {
					tx.Rollback()
				} else if err != nil {
					if rbErr := tx.Rollback; rbErr != nil {
						fmt.Println("Error Rollback: ", rbErr)
					}
				} else {
					cErr := tx.Commit()
					if cErr != nil {
						fmt.Println("Error Commit: ", cErr)
					} else {
						fmt.Println("Delete Success from DB, ID File: ", itemData.ItemID)
					}
				}

			} else {
				fmt.Println("File: ", filepath)
				fmt.Println("Event: ", event)
			}

		case err := <-watcher.Errors:
			fmt.Println("Error: ", err)
		}
	}
}
