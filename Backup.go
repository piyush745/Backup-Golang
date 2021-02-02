package main

import (
    "encoding/json"
    "fmt"
    "strings"
	  "io"
	 "io/ioutil"
   "os"
	 "path"
	"path/filepath"
  "archive/zip"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/JamesStewy/go-mysqldump"
)

type Contents struct {
  Contents []Content `json:"contents"`
}

type Content struct {
  SourceDirectory string `json:"source-directory"`
  SourceFile string `json:"source-file"`
  BackupDirectory string `json:"destination-directory"`
  ZipDirectory string `json:"zip-directory"`
  SQL SQL `json:"sql"`
}

type SQL struct {
  Username string `json:"username"`
  Password  string `json:"password"`
  DBname string `json:"dbname"`
}

func main() {
    
    Data, err := ioutil.ReadFile("./example.json")
    if err != nil {
      fmt.Print(err.Error())
    }
    var contents Contents
    err2 := json.Unmarshal(Data, &contents)
    if err2 != nil {
    	fmt.Println("Error")
    	fmt.Println(err2.Error())
    }
    backuppath := contents.Contents[0].BackupDirectory
    zippath := contents.Contents[0].ZipDirectory
    set_backup_directory(backuppath)

    for i := 0; i < len(contents.Contents); i++ {
      dirstr := contents.Contents[i].SourceDirectory
      if dirstr != "" {
      dirlist := strings.Split(contents.Contents[i].SourceDirectory,",")
      for j := 0; j < len(dirlist); j++ {
        dirnames := strings.Split(dirlist[j],"/")
        dirname := dirnames[len(dirnames)-1]
        fmt.Println(dirname)
        absolutePath :=backuppath+dirname
        fmt.Println(absolutePath)
        Dir(dirlist[i],absolutePath)
      }  
      }
       
  }

  for i := 0; i < len(contents.Contents); i++ {
    filestr := contents.Contents[i].SourceFile
    if filestr != "" {
      filelist :=strings.Split(contents.Contents[i].SourceFile,",")
    fmt.Println(len(filelist))
    for j := 0; j < len(filelist); j++ {
      fmt.Println("Entered the loop")
      filenames := strings.Split(filelist[j],"/")
      filename := filenames[len(filenames)-1]
      fmt.Println(filename)
      absolutePath :=backuppath+filename
      fmt.Println(absolutePath)
      File(filelist[i],absolutePath)
    }
  }
}
 
  username := contents.Contents[0].SQL.Username
  password := contents.Contents[0].SQL.Password
  dbname := contents.Contents[0].SQL.DBname
  sql_dump(username,password,dbname,backuppath)
  zipit(backuppath,zippath)
}

func set_backup_directory(path string){
  err := os.RemoveAll(path) 
    if err != nil { 
        fmt.Println(err)
    } 
  if _, err := os.Stat(path); os.IsNotExist(err) {
    os.Mkdir(path, 0755)
}
}

func sql_dump(username,password,dbname,backuppath string)  {
  connectionpath := username+":"+password+"@tcp(127.0.0.1:3306)/"+dbname
  db, err := sql.Open("mysql", connectionpath)
	if err != nil {
        panic(err.Error())
  }
  defer db.Close()

  dumper, err := mysqldump.Register(db, backuppath, "sqldump")
    if err != nil {
    	fmt.Println("Error registering databse:", err)
    	return
    }

    // Dump database to file
    _,err = dumper.Dump()
    if err != nil {
    	fmt.Println("Error dumping:", err)
    	return
    }
    fmt.Printf("File is saved ")

    // Close dumper and connected database
    dumper.Close()
}

  func zipit(source, target string) error {
    zipfile, err := os.Create(target)
    if err != nil {
      return err
    }
    defer zipfile.Close()
  
    archive := zip.NewWriter(zipfile)
    defer archive.Close()
  
    info, err := os.Stat(source)
    if err != nil {
      return nil
    }
  
    var baseDir string
    if info.IsDir() {
      baseDir = filepath.Base(source)
    }
  
    filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
      if err != nil {
        return err
      }
  
      header, err := zip.FileInfoHeader(info)
      if err != nil {
        return err
      }
  
      if baseDir != "" {
        header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
      }
  
      if info.IsDir() {
        header.Name += "/"
      } else {
        header.Method = zip.Deflate
      }
  
      writer, err := archive.CreateHeader(header)
      if err != nil {
        return err
      }
  
      if info.IsDir() {
        return nil
      }
  
      file, err := os.Open(path)
      if err != nil {
        return err
      }
      defer file.Close()
      _, err = io.Copy(writer, file)
      return err
    })
  
    return err
  }
  

  func File (src,dst string) error {
    in, err := os.Open(src)
		if err != nil {
			return err
		}
		defer in.Close()
	
		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer out.Close()
	
		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}
		return out.Close()
  }
  
  func Dir(src string, dst string) error {
    var err error
    var fds []os.FileInfo
    var srcinfo os.FileInfo
  
  
  dir, err := ioutil.ReadDir(dst)
      for _, d := range dir {
          os.RemoveAll(path.Join([]string{"tmp", d.Name()}...))
      }
    if srcinfo, err = os.Stat(src); err != nil {
      return err
    }
  
    if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
      return err
    }
  
    if fds, err = ioutil.ReadDir(src); err != nil {
      return err
    }
    for _, fd := range fds {
      srcfp := path.Join(src, fd.Name())
      dstfp := path.Join(dst, fd.Name())
  
      if fd.IsDir() {
        if err = Dir(srcfp, dstfp); err != nil {
          fmt.Println(err)
        }
      } else {
        if err = File(srcfp, dstfp); err != nil {
          fmt.Println(err)
        }
      }
    }
    return nil
  }
  
  
