package arcpics

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func picturesAndDatabaseDirectories(args []string) (string, string) {
	picturesDirName := defaultPicturesDirName
	databaseDirName := GetDatabaseDirName()

	if len(args) < 1 {
		return picturesDirName, databaseDirName

	}
	if len(args) >= 1 {
		picturesDirName = args[0]

	}
	if len(args) > 1 {
		databaseDirName = args[1]
	}
	return picturesDirName, databaseDirName
}
func DirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// returns exists and isFile
func FileExists(filename string) bool {
	_, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	return !os.IsNotExist(err)
}
func GetDatabaseDirName() string {
	var userHomeDir string = "."
	var err error
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting UserHomeDir: " + err.Error())
	}
	databaseDirName := filepath.Join(userHomeDir, dotDefaultName)
	ok := DirExists(databaseDirName)
	if !ok {
		err = os.Mkdir(databaseDirName, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s :%s \n", databaseDirName, err.Error())
		} else {
			fmt.Printf("Directory %s created.\n", databaseDirName)
		}
	}
	return databaseDirName
}
func insertSystemLabelValue(db *bolt.DB, value string) error {
	//return insertDbValue(db, SYSTEM_BUCKET, keyBytes)
	keyBytes := []byte(INIT_LABEL_KEY)
	err := db.Update(func(tx *bolt.Tx) error {
		foundBytes := tx.Bucket([]byte(SYSTEM_BUCKET)).Get(keyBytes)
		if len(foundBytes) == 0 {
			err := tx.Bucket([]byte(SYSTEM_BUCKET)).Put(keyBytes, []byte(value))
			if err != nil {
				return fmt.Errorf("bucket %s - Could not insert entry: %v", SYSTEM_BUCKET, err)
			}
		}
		return nil
	})
	return err
}
func insert2System_KeyValueStrings(db *bolt.DB, key, value string) error {
	return insert2System_KeyValue(db, []byte(key), []byte(value))
}
func insert2System_KeyValue(db *bolt.DB, keyBytes, valueBytes []byte) error {
	return insert2bucket_KeyValue(db, []byte(SYSTEM_BUCKET), keyBytes, valueBytes)
}
func insert2MountDirNow(db *bolt.DB, dir string) {
	insert2bucket_KeyValue(db, []byte(MOUNTS_BUCKET), []byte(dir), []byte(time.Now().Format(time.DateTime)))
}
func lastMountedDir(db *bolt.DB) string {
	mountedDirsInPast := ArcpicsAllKeys(db, MOUNTS_BUCKET)
	latestTime := ""
	latestDir := ""
	for _, dir := range mountedDirsInPast {
		t := GetDbValue(db, MOUNTS_BUCKET, dir)
		if t > latestTime {
			latestTime = t
			latestDir = dir
		}
	}
	return latestDir
}
func insert2bucket_KeyValue(db *bolt.DB, bucket, keyBytes, valueBytes []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(bucket).Put(keyBytes, valueBytes)
		if err != nil {
			return fmt.Errorf("bucket:%s, key:%s - Could not insert value:%v", string(bucket), string(keyBytes), err)
		}
		return nil
	})
	return err
}
func getSystemLabelSummary(label string) string {
	dbDir := GetDatabaseDirName()
	databaseName := filepath.Join(dbDir, defaultNameDash+label+".db")

	var db *bolt.DB
	db, err := bolt.Open(databaseName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	key := fmt.Sprintf(LABEL_SUMMARY_fmt, label)
	keyBytes := []byte(key)
	var foundBytes []byte
	db.View(func(tx *bolt.Tx) error {
		foundBytes = tx.Bucket([]byte(SYSTEM_BUCKET)).Get(keyBytes)
		return nil
	})
	return string(foundBytes)
}
func insertNewBucket(db *bolt.DB, bucket []byte) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		panic(err)
	}
}

////////////////////////
// Exported functions //
////////////////////////

func DbLabel(archiveDir string) (string, error) {
	nameStart := defaultNameDashLabelDot
	label := ""
	files, err := os.ReadDir(archiveDir)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), nameStart) {
			count++
			label = f.Name()[len(nameStart):]
		}
	}
	if count == 0 {
		return label, fmt.Errorf("there is no file %s* at directory %s - should be e.g. %s001", nameStart, archiveDir, nameStart)
	} else if count > 1 {
		return label, fmt.Errorf("unexpected number files %s* at directory %s", nameStart, archiveDir)
	}
	if len(label) < 1 {
		return label, fmt.Errorf("there is not at least one character label after dot 'arcpics-db-label.'")
	}
	return label, nil
}

func PutDbValueHttpReqDir(db *bolt.DB, bucket []byte, keyStr string, jd *JdirType) error {
	var err error
	var jdBytes []byte
	jdBytes, err = json.Marshal(jd)
	if err != nil {
		return err
	}
	err = PutDbValue(db, bucket, keyStr, string(jdBytes))
	if err != nil {
		return err
	}
	return err
}

func PutDbValue(db *bolt.DB, bucket []byte, key, value string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put([]byte(key), []byte(value))
	})
	return err
}

func GetDbValue(db *bolt.DB, bucket []byte, key string) string {
	var valueBytes []byte
	_ = db.View(func(tx *bolt.Tx) error {
		valueBytes = tx.Bucket([]byte(bucket)).Get([]byte(key))
		return nil
	})
	return string(valueBytes)
}

func AssignPicturesDirectoryWithDatabase(varArgs ...string) (ArcpicsFS, *bolt.DB, error) {
	var arcFS ArcpicsFS
	picturesDirName, databaseDirName := picturesAndDatabaseDirectories(varArgs)
	label, err := DbLabel(picturesDirName)
	if err != nil {
		return arcFS, nil, err
	}
	databaseName := filepath.Join(databaseDirName, defaultNameDash+label+".db")
	dbDidExist := FileExists(databaseName)

	// Open the database data file. It will be created if it doesn't exist.
	var db *bolt.DB
	db, err = bolt.Open(databaseName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	if dbDidExist {
		// check INIT key
		s := GetDbValue(db, SYSTEM_BUCKET, INIT_LABEL_KEY)
		if s != label {
			return arcFS, db, fmt.Errorf("init value  %s at DB %s doesn't match %s at dir %s", s, databaseName, label, picturesDirName)
		}
	} else {
		insertNewBucket(db, SYSTEM_BUCKET) // insert SYSTEM bucket just once

		// insert init key into SYSTEM bucket just once
		err = insertSystemLabelValue(db, label)
		if err != nil {
			panic(err)
		}
		insertNewBucket(db, FILES_BUCKET)  // insert FILES bucket just once
		insertNewBucket(db, MOUNTS_BUCKET) // insert MOUNTD bucket just once
		mountDir := picturesDirName
		if mountDir, err = filepath.Abs(picturesDirName); err != nil {
			fmt.Printf("error making abs mountDir: %s\n", err.Error())
		}
		insert2MountDirNow(db, mountDir)
	}
	picturesDirName = strings.TrimSuffix(picturesDirName, "/")
	arcFS, err = OpenArcpicsFS(picturesDirName)
	if err != nil {
		fmt.Println("error: " + err.Error())
		os.Exit(1)
	}

	return arcFS, db, nil
}

func LabeledDatabase(label string, varArgs ...string) (*bolt.DB, error) {
	databaseDirName := GetDatabaseDirName()

	if len(varArgs) >= 1 {
		databaseDirName = varArgs[0]
	}

	databaseName := filepath.Join(databaseDirName, "arcpics-"+label+".db")
	dbExists := FileExists(databaseName)
	if !dbExists {
		return nil, fmt.Errorf("db file %s not found", databaseName)
	}

	// Open the database file. It will be created if it doesn't exist.
	var db *bolt.DB
	var err error
	db, err = bolt.Open(databaseName, 0400, nil) //Read only
	if err != nil {
		return db, err
	}

	// check INIT key
	s := GetDbValue(db, SYSTEM_BUCKET, INIT_LABEL_KEY)
	if s != label {
		return db, fmt.Errorf("init value  %s at DB %s doesn't match label %s", s, databaseName, label)
	}

	return db, nil
}
func CreateLabelFile(dirName, newLabel string) error {
	fname := filepath.Join(dirName, defaultNameDashLabelDot+newLabel)
	if FileExists(fname) {
		return fmt.Errorf("file %s already exists", fname)
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	if _, err = io.WriteString(f, fmt.Sprintf("label is defined by an extension of this file: %s\n", newLabel)); err != nil {
		return err
	}
	fmt.Printf("Label file %s was created\n", fname)
	return nil
}

/*
func filesInDir(d string) ([]fs.DirEntry, error) {
	var files []fs.DirEntry
	var err error
	if files, err = os.ReadDir(d); err != nil {
		return nil, err
	}
	onlyFiles := make([]fs.DirEntry, 0)
	for _, file := range files {
		if !file.IsDir() {
			onlyFiles = append(onlyFiles, file)
		}
	}

	for _, f := range onlyFiles {
		info, _ := f.Info()
		fmt.Printf("    fs.DirEntry: %v %v\n", f.Name(), info.ModTime())
	}
	fmt.Printf("----\n")

	return onlyFiles, nil
}

func FilesInDirXXX(d string) (JdirType, error) {
	var jd JdirType
	jd.Files = make([]JfileType, 0)
	jd.Description = "a description..."
	jd.Location = "a secret location..."
	var files []fs.DirEntry
	var err error
	if files, err = filesInDir(d); err != nil {
		return jd, err
	}

	for _, f := range files {
		info, _ := f.Info()
		var file JfileType
		file.Name = info.Name()
		file.Time = info.ModTime().Format("2006-01-02_15:04:05")
		file.Comment = "my comment, realy"
		jd.Files = append(jd.Files, file)
	}
	return jd, nil
}

func CreateDirJson(d string, dirFiles JdirType) error {
	_, _ = filesInDir(d)
	jsonBytes, err := json.Marshal(dirFiles)
	if err != nil {
		return err
	}
	fname := filepath.Join(d, defaultNameJson+"-new")
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	_, err = f.Write(jsonBytes) //lenght ommited
	if err != nil {
		return err
	}
	f.Close()
	return nil
}
func UpdateDirJson(d string) error {
	fjson_new := filepath.Join(d, defaultNameJson+"-new")
	if _, err := os.Stat(fjson_new); err != nil {
		return nil // there is no new file
	}
	fjson := filepath.Join(d, defaultNameJson)
	var fjson_FileInfo os.FileInfo
	var err error
	if fjson_FileInfo, err = os.Stat(fjson); err != nil {
		return nil // there is now current file
	}
	timeExtension := fjson_FileInfo.ModTime().Format("--2006-01-02_15-04-05") // Mon Jan 2 15:04:05 -0700 MST 2006
	fjson_time := filepath.Join(d, defaultNameJson+timeExtension)
	if err := os.Rename(fjson, fjson_time); err != nil {
		return err
	}
	if err := os.Rename(fjson_new, fjson); err != nil {
		return err
	}
	return nil
}
*/
