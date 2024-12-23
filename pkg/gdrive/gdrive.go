package gdrive

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/momokii/ss-watcher/pkg/utils"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GDrive interface {
	GetService() *drive.Service
	CheckFolderExist(folderName string, parentId string) (string, error)
	CreateFolder(folderName string, parentId string) (string, error)
	CheckExistOrCreateFolderSSDaily(parentFolderCheckId string) (string, error)
	UploadFileDrive(filename, filepath, mimeType, parentFolderId string) (*drive.File, error)
	DeleteFileDrive(id string) error
	NewUserPermission(base_gdrive_folder_id, user_email string) (string, error)
	DeleteUserPermission(permission_id string) error
}

type gdrive struct {
	Service *drive.Service
}

func NewGDrive(service_account_path string) GDrive {
	ctx := context.Background()

	// Replace with the path to your service account JSON file.
	serviceAccountFile := service_account_path

	// Initialize the Drive service using the service account file.
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(serviceAccountFile), option.WithScopes(drive.DriveScope))
	if err != nil {
		fmt.Println("Error creating Drive service: ", err)
	} else {
		fmt.Println("Drive service connected successfully")
	}

	return &gdrive{
		Service: srv,
	}
}

func (d *gdrive) GetService() *drive.Service {
	return d.Service
}

func (d *gdrive) CheckFolderExist(folderName string, parentId string) (string, error) {
	// gdrive query to check folder

	query := fmt.Sprintf("name contains '%s' and mimeType='application/vnd.google-apps.folder'", folderName)

	if parentId != "" {
		query += fmt.Sprintf(" and '%s' in parents", parentId)
	}

	query += " and trashed=false"

	fileList, err := d.Service.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return "", fmt.Errorf("Error checking folder: %v", err)
	}

	if len(fileList.Files) > 0 {
		fmt.Printf("Folder %s already exists with id: %s \n", folderName, fileList.Files[0].Id)
		return fileList.Files[0].Id, nil
	}

	fmt.Println("Folder does not exist, creating folder...")
	return "", nil
}

func (d *gdrive) CreateFolder(folderName string, parentId string) (string, error) {
	// define folder
	folder := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}

	// if parent provided, set it
	rootParent := "root"

	folder.Parents = []string{parentId}
	if parentId == "" {
		folder.Parents = []string{rootParent}
	}

	// create folder
	createdFolder, err := d.Service.Files.Create(folder).Fields("id", "name").Do()
	if err != nil {
		return "", fmt.Errorf("Error creating folder: %v", err)
	}

	fmt.Printf("Folder %s created with id: %s \n", folderName, createdFolder.Id)
	return createdFolder.Id, nil
}

func (d *gdrive) CheckExistOrCreateFolderSSDaily(parentFolderCheckId string) (string, error) {
	nameFolder := "SS_" // prefix folder name
	dateFolder := time.Now().Format("2006-01-02")
	randomCode := utils.RandomString(5)
	checkFolderName := nameFolder + dateFolder
	newFolderName := checkFolderName + "_" + randomCode

	folderId, err := d.CheckFolderExist(checkFolderName, parentFolderCheckId) // id "test" folder
	if err != nil {
		fmt.Println("Error Check Folder Exist: ", err)
		return "", err
	}

	if folderId == "" {
		fmt.Println("Creating Daily folder:")
		folderId, err = d.CreateFolder(newFolderName, parentFolderCheckId)
		if err != nil {
			fmt.Println("Error Create Folder: ", err)
			return "", err
		}
	}

	return folderId, nil
}

func (d *gdrive) UploadFileDrive(filename, filepath, mimeType, parentFolderId string) (*drive.File, error) {

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("Error Open File: %v", err)
	}
	defer file.Close()

	fileMetadata := &drive.File{
		Name:     filename,
		MimeType: mimeType,
	}

	if parentFolderId != "" {
		fileMetadata.Parents = []string{parentFolderId}
	}

	fileUpload, err := d.Service.Files.Create(fileMetadata).Media(file).Do()
	if err != nil {
		return nil, fmt.Errorf("Error Upload File: %v", err)
	}

	return fileUpload, nil
}

func (d *gdrive) DeleteFileDrive(id string) error {
	file, err := d.Service.Files.Get(id).Fields("mimeType").Do()
	if err != nil {
		return fmt.Errorf("Error Get File: %v", err)
	}

	// this function cannt delete folder
	if file.MimeType == "application/vnd.google-apps.folder" {
		return fmt.Errorf("Cannot delete folder")
	}

	// delete file
	if err := d.Service.Files.Delete(id).Do(); err != nil {
		return fmt.Errorf("Error Delete File: %v", err)
	}
	fmt.Println("File deleted with id: ", id)

	return nil
}

func (d *gdrive) NewUserPermission(base_gdrive_folder_id, user_email string) (string, error) {
	perm := &drive.Permission{
		Type:         "user",
		Role:         "writer",
		EmailAddress: user_email,
	}

	// give permission to owner as writer so the service account still can access the folder
	permission, err := d.Service.Permissions.Create(base_gdrive_folder_id, perm).Do()
	if err != nil {
		fmt.Println("Error Transfer Ownership: ", err)
		// delete folder if error on giving ownership
		if err := d.Service.Files.Delete(base_gdrive_folder_id).Do(); err != nil {
			return "", fmt.Errorf("Error Canceling Create Base Folder: %v", err)
		}

		return "", fmt.Errorf("Cancelling Create Base Folder Success")
	}

	// return permission id
	return permission.Id, nil
}

func (d *gdrive) DeleteUserPermission(permission_id string) error {
	// delete permission
	if err := d.Service.Permissions.Delete(permission_id, permission_id).Do(); err != nil {
		return fmt.Errorf("Error Delete Permission: %v", err)
	}

	return nil
}
