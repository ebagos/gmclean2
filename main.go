package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

type Config struct {
	Dirs []string `json:"dirs"`
}

type File struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Date string `json:"date"`
	Hash string `json:"hash"`
}

type DirData struct {
	Files []File `json:"files"`
}

// 引数を解析する
func parseArgs() (string, error) {
	if len(os.Args) < 2 {
		return "", nil
	}
	return os.Args[1], nil
}

// ファイルのハッシュ値を計算する
func calcHash(path string) (string, error) {
	// ファイルを開く
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// ハッシュ値を計算する
	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// ハッシュ値を16進数文字列に変換する
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// results.json を取得する
func getDirData(path string) (DirData, error) {
	// ファイルを開く
	file, err := os.Open(path)
	if err != nil {
		return DirData{}, err
	}
	defer file.Close()

	// ファイルを読み込む
	var dirData DirData
	if err := json.NewDecoder(file).Decode(&dirData); err != nil {
		return DirData{}, err
	}

	// ファイルの情報を返す
	return dirData, nil
}

// 与えられたファイルが DirDatas.json に記録されているか確認する
// 一致するファイル名があった場合、サイズと更新日付を確認し、一致したらそのままリターンする
// ファイル名が一致するがサイズや更新日付が一致しない場合、ファイル情報を更新する

// 一致するファイル名がなかった場合、ファイル情報を獲得し、DirDatas.json に追加する
func checkFile(dirData DirData, path string) (DirData, error) {
	// 最初にファイルのサイズと更新日付を確認する
	info, err := os.Stat(path)
	if err != nil {
		return DirData{}, err
	}
	// ファイル名が一致するファイルがあるか確認する
	for i, file := range dirData.Files {
		if file.Path == path {
			if file.Size == info.Size() && file.Date == info.ModTime().String() {
				// ファイル情報をそのまま返す
				return dirData, nil
			} else {
				// ファイルの情報を更新する
				dirData.Files[i].Date = info.ModTime().String()
				dirData.Files[i].Hash, err = calcHash(path)
				return dirData, nil
			}
		}
	}
	hash, err := calcHash(path)
	if err != nil {
		return DirData{}, err
	}
	// ファイル情報を追加する
	dirData.Files = append(dirData.Files, File{
		Path: path,
		Size: info.Size(),
		Date: info.ModTime().String(),
		Hash: hash,
	})
	return dirData, nil
}

// Results.jsonにあるファイルのうち、存在しないファイルをResults.jsonから削除する
func removeFile(dirData DirData, path string) (DirData, error) {
	for i, file := range dirData.Files {
		if _, err := os.Stat(file.Path); err != nil {
			dirData.Files = append(dirData.Files[:i], dirData.Files[i+1:]...)
		}
	}
	return dirData, nil
}

// results.jsonがなかった場合、新規に作成する
func createDirData(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	dirData := DirData{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		info, err := os.Stat(dir + "/" + f.Name())
		if err != nil {
			return err
		}
		hash, err := calcHash(dir + "/" + f.Name())
		if err != nil {
			return err
		}

		dirData.Files = append(dirData.Files, File{
			Path: dir + "/" + f.Name(),
			Size: info.Size(),
			Date: info.ModTime().String(),
			Hash: hash,
		})
		// DirDataをJSONに変換する
		dirDataJSON, err := json.Marshal(dirData)
		if err != nil {
			return err
		}
		// results.jsonを作成する
		if err := os.WriteFile(dir+"/results.json", dirDataJSON, 0666); err != nil {
			return err
		}
	}
	return nil
}

// DirDatasをDirDatas.jsonに書き込む
func writeDirData(dirData DirData, path string) error {
	// DirDataをJSONに変換する
	dirDataJSON, err := json.Marshal(dirData)
	if err != nil {
		return err
	}
	// results.jsonを作成する
	if err := os.WriteFile(path, dirDataJSON, 0666); err != nil {
		return err
	}
	return nil
}

// results.jsonからFileハッシュ値が同じものを削除する
func removeSameHash(all []DirData) {
	// allをfilesに変換する
	files := []File{}
	for _, dirData := range all {
		files = append(files, dirData.Files...)
	}
	// filesをハッシュ値でソートする
	// ハッシュ値が同じファイルが連続するようになる
	// 連続するハッシュ値が同じファイルのうち最新の更新日付のものを残して削除する
	// 連続するハッシュ値が同じファイルがなくなるまで繰り返す
	for {
		// ハッシュ値でソートする
		sort.Slice(files, func(i, j int) bool {
			return files[i].Hash < files[j].Hash
		})
		// 連続するハッシュ値が同じファイルのうち最新の更新日付のものを残して削除する
		removed := false
		fmt.Println(files)
		for i := 0; i < len(files)-1; i++ {
			if files[i].Hash == files[i+1].Hash {
				fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!\n", files[i].Path, files[i+1].Path)
				if files[i].Date < files[i+1].Date {
					files = append(files[:i], files[i+1:]...)
				} else {
					files = append(files[:i+1], files[i+2:]...)
				}
				removed = true
				break
			}
		}
		// 連続するハッシュ値が同じファイルがなくなったら終了する
		if !removed {
			break
		}
	}
}

// 単一ディレクトリの処理
func processDir(dir string) error {
	// results.jsonがなかった場合、新規に作成する
	if _, err := os.Stat(dir + "/results.json"); err != nil {
		if err := createDirData(dir); err != nil {
			return nil
		}
	}

	// results.jsonを取得する
	dirData, err := getDirData(dir + "/results.json")
	if err != nil {
		return err
	}

	// ディレクトリ内のファイルを確認する
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// ファイルの存在を確認する
		if _, err := os.Stat(dir + "/" + f.Name()); err != nil {
			// ファイルが存在しない場合、results.jsonから削除する
			dirData, err = removeFile(dirData, dir+"/"+f.Name())
			if err != nil {
				return err
			}
			continue
		}
		// ファイルが存在する場合、results.jsonに記録されているか確認する
		dirData, err = checkFile(dirData, dir+"/"+f.Name())
		if err != nil {
			return err
		}
	}

	// results.jsonを更新する
	if err := writeDirData(dirData, dir+"/results.json"); err != nil {
		return err
	}

	return nil
}

func main() {
	// 引数を解析する
	configPath, err := parseArgs()
	if err != nil {
		panic(err)
	}
	if configPath == "" {
		configPath = "./config.json"
	}

	// config.json を読み込む
	configFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	defer configFile.Close()
	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		panic(err)
	}

	// 各ディレクトリのresults.jsonを更新する
	for _, dir := range config.Dirs {
		if err := processDir(dir); err != nil {
			panic(err)
		}
	}

	// 全てのresults.jsonからFileハッシュ値が同じものを削除する
	all := []DirData{}
	for _, dir := range config.Dirs {
		DirData, err := getDirData(dir + "/results.json")
		if err != nil {
			panic(err)
		}
		all = append(all, DirData)
	}
	removeSameHash(all)
}
