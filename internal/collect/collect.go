package collect

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"

	"github.com/tidwall/gjson"
)

type User struct {
	Name     string
	Pictures []string
}

func Scrape(fastupdate bool, profiles []string, threads int) (users []User) {
	var wg sync.WaitGroup
	scrapedata := make(chan User, len(profiles))
	ch := make(chan string, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for profile := range ch {
				GetScarpeData(scrapedata, fastupdate, profile)
			}
		}()
	}

	for _, profile := range profiles {
		ch <- profile
	}
	close(ch)
	wg.Wait()
	close(scrapedata)
	for user := range scrapedata {
		users = append(users, user)
	}

	return users
}

func GetScarpeData(scrapedata chan User, fastupdate bool, profile string) (users []User) {
	re := regexp.MustCompile(`__PRELOADED_STATE__ = (.*)<\/script>`)
	fast := fastupdate
	if _, err := os.Stat(profile); err != nil {
		fast = false
	}
	urladdr := fmt.Sprintf("https://vsco.co/%s/gallery", profile)
	var more bool
	more = true

	resp, err := http.Get(urladdr)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		fmt.Printf("Profile %s does not exists\n", profile)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	data := re.FindStringSubmatch(string(body))
	if len(data) > 0 {
		user := User{Name: profile}
		id := gjson.Get(data[1], fmt.Sprintf("sites.siteByUsername.%s.site.id", profile))
		token := gjson.Get(data[1], "users.currentUser.tkn")
		nextcursor := gjson.Get(data[1], fmt.Sprintf("medias.bySiteId.%s.nextCursor", id)).String()
		if nextcursor == "" || fast {
			more = false
		}
		result := gjson.Get(data[1], "entities.images.@keys")
		if len(result.Array()) == 0 {
			fmt.Printf("Profile %s pictures gone\n", profile)
			return
		}
		for _, key := range result.Array() {
			result := gjson.Get(data[1], fmt.Sprintf("entities.images.%s", key))
			user.Pictures = append(user.Pictures, result.String())
		}
		for more {
			urlmore := fmt.Sprintf("https://vsco.co/api/3.0/medias/profile?site_id=%s&limit=14&cursor=%s", id, url.QueryEscape(nextcursor))
			client := http.Client{}
			req, err := http.NewRequest("GET", urlmore, nil)
			if err != nil {
				panic(err)
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			result = gjson.Get(string(body), "media.#.image")
			for _, image := range result.Array() {
				user.Pictures = append(user.Pictures, image.String())
			}

			value := gjson.Get(string(body), "next_cursor")
			if value.Exists() {
				nextcursor = value.String()
			} else {
				more = false
			}
		}
		scrapedata <- user
	} else {
		fmt.Printf("Profile %s pictures are gone\n", profile)
		return
	}
	return
}
