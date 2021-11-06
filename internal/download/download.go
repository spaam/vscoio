package download

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/spaam/vscoio/internal/utils"

	"github.com/schollz/progressbar/v3"
	"github.com/spaam/vscoio/internal/collect"
	"github.com/tidwall/gjson"
	"github.com/ulikunitz/xz"
)

func Download(nometadata bool, fastupdate bool, threads int, Users []collect.User) int {
	var total int
	for _, user := range Users {
		var dl_list []string
		name := user.Name
		_, err := os.Stat(name)
		if err != nil {
			os.Mkdir(name, os.ModePerm)
		}
		files, _ := ioutil.ReadDir(name)
		for _, pic := range user.Pictures {
			if fastupdate && !stringInSlice(utils.Filename(utils.GetUploadDate(pic)), files) {
				dl_list = append(dl_list, pic)

			} else if !fastupdate {
				dl_list = append(dl_list, pic)
			}
		}

		if len(dl_list) == 0 {
			continue
		}

		total += len(dl_list)
		bar := progressbar.Default(int64(len(dl_list)), name)
		var wg sync.WaitGroup
		ch := make(chan string, threads)

		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for d := range ch {
					getPictures(nometadata, d, user)
				}
			}()
		}
		for _, picture := range dl_list {
			bar.Add(1)
			ch <- picture
		}
		close(ch)
		wg.Wait()
		bar.Close()
	}
	return total
}

func getPictures(nometadata bool, picture string, user collect.User) {
	if _, err := os.Stat(fmt.Sprintf("%s/%s", user.Name, utils.Filename(utils.GetUploadDate(picture)))); err == nil {
		fmt.Printf("%s file exists, skipping\n", fmt.Sprintf("%s/%s", user.Name, utils.Filename(utils.GetUploadDate(picture))))
		return
	}
	resp, err := http.Get(fmt.Sprintf("https://%s", getResponsiveurl(picture)))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fd, err := os.Create(fmt.Sprintf("%s/%s", user.Name, utils.Filename(utils.GetUploadDate(picture))))
	if err != nil {
		fmt.Println(err)
		return
	}
	fd.Write(body)
	fd.Close()
	if !nometadata {
		fd2, _ := os.Create(fmt.Sprintf("%s/%s.json.xz", user.Name, utils.Filename(utils.GetUploadDate(picture))))
		w, _ := xz.NewWriter(fd2)
		w.Write([]byte(picture))
		w.Close()
		fd2.Close()
	}
}

func getResponsiveurl(picture string) string {
	value := gjson.Get(picture, "responsiveUrl")
	if value.Exists() {
		return value.String()
	}
	return gjson.Get(picture, "responsive_url").String()
}

func stringInSlice(a string, list []fs.FileInfo) bool {
	for _, b := range list {
		if b.Name() == a {
			return true
		}
	}
	return false
}
