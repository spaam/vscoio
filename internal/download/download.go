package download

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/schollz/progressbar/v3"
	"github.com/spaam/vscoio/internal/collect"
	"github.com/tidwall/gjson"
	"github.com/ulikunitz/xz"
)

func Download(nometadata bool, threads int, Users []collect.User) {
	for _, user := range Users {
		var dl_list []string
		name := user.Name
		_, err := os.Stat(name)
		if err != nil {
			os.Mkdir(name, os.ModePerm)
		}
		files, _ := ioutil.ReadDir(name)

		for _, pic := range user.Pictures {
			if !stringInSlice(filename(getUploadDate(pic)), files) {
				dl_list = append(dl_list, pic)
			}
		}
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
	}
}

func getPictures(nometadata bool, picture string, user collect.User) {
	resp, err := http.Get(fmt.Sprintf("https://%s", getResponsiveurl(picture)))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fd, err := os.Create(fmt.Sprintf("%s/%s", user.Name, filename(getUploadDate(picture))))
	if err != nil {
		fmt.Println(err)
		return
	}
	fd.Write(body)
	fd.Close()
	if !nometadata {
		fd2, _ := os.Create(fmt.Sprintf("%s/%s.json.xz", user.Name, filename(getUploadDate(picture))))
		w, _ := xz.NewWriter(fd2)
		w.Write([]byte(picture))
		w.Close()
		fd2.Close()
	}
}

func getUploadDate(picture string) int64 {
	value := gjson.Get(picture, "uploadDate")
	if value.Exists() {
		return value.Int() / 1000
	}
	return gjson.Get(picture, "upload_date").Int() / 1000
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

func filename(timestamp int64) string {
	tm := time.Unix(timestamp, 0)
	s, err := strftime.Format("%Y-%m-%d_%H-%M-%S", tm)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s_UTC.jpg", s)
}
