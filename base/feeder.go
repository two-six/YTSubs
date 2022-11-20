package base

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"html"
)

const RSS_LINK = "https://www.youtube.com/feeds/videos.xml?channel_id="

type Feed struct {
	videos []Video
	feed string
}

type Video struct {
	link string
	title string
	author string
	thumbnail string
	id string
}

func Fetch(channel_id string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s", RSS_LINK, channel_id))
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		bodyString := string(bodyBytes)
		return bodyString, nil
	}
	return "", err
}

func Extract_videos(feed string) []Video {
	re_title, _ := regexp.Compile("<media:title>(.*?)</media:title>")
	re_link, _ := regexp.Compile(`<media:content url="(.*?)"`)
    re_author, _ := regexp.Compile(`<author>\n\s*<name>(.*?)</name>`)
	re_thumbnail, _ := regexp.Compile(`<media:thumbnail url="(.*?)"`)
	re_id, _ := regexp.Compile(`<yt:videoId>(.*?)</yt:videoId>`)
	var res []Video
	sections := strings.Split(feed, "</entry>")
	for _, section := range(sections[:len(sections)-1]) {
		var vid Video
		vid.title = html.UnescapeString(re_title.FindStringSubmatch(section)[1])
		vid.link = html.UnescapeString(re_link.FindStringSubmatch(section)[1])
		vid.author = html.UnescapeString(re_author.FindStringSubmatch(section)[1])
		vid.thumbnail = html.UnescapeString(re_thumbnail.FindStringSubmatch(section)[1])
		vid.id = html.UnescapeString(re_id.FindStringSubmatch(section)[1])
		res = append(res, vid)
	}
	return res
}

func Get_instances_list() ([]string, error) {
	resp, err := http.Get("https://api.invidious.io/instances.json")
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		bodyString := string(bodyBytes)
		re_links, _ := regexp.Compile(`uri":"(.*?)"`)
		links := re_links.FindAllStringSubmatch(bodyString, -1)
		var res []string
		for _, link := range(links) {
			if !(strings.Contains(link[1], ".i2p") || strings.Contains(link[1], ".onion")) {
				link := strings.Replace(link[1], "https://", "", 1)
				res = append(res, link)
			}
		}
		return res, nil

	}
	return nil, err
}

func Extract_channel_id(link string) (string, error) {
	link = strings.Replace(link, "https://", "", 1)
	link = strings.Replace(link, "http://", "", 1)
	link = strings.Replace(link, "www.", "", 1)
	instances, err := Get_instances_list()
	if err != nil {
		return "", err
	}
	for _, instance := range(instances) {
		resp, _ := http.Get(fmt.Sprintf("https://www.%s", instance))
		if resp.StatusCode != http.StatusOK {
			continue
		}
		channel_id, err := extract_channel_id_instance(link, instance)
		if err != nil {
			return "", err
		}
		if strings.Contains(channel_id, "/user/") {
			channel_id, err := extract_channel_id_instance(strings.Replace(link, "@", "/c/", 1), instance)
			if err != nil {
				return "", err
			}
			return channel_id, nil
		}
		return channel_id, nil
	}
	return "", fmt.Errorf("Couldn't extract channel id from instances")
}

func extract_channel_id_instance(link string, instance string) (string, error) {
	link = strings.Replace(link, "youtube.com", fmt.Sprintf("https://www.%s", instance), 1)
	link = strings.Replace(link, "@", "user/", 1)
	resp, err := http.Get(link)
	if err != nil {
		return "", err
		}
	return strings.Replace(resp.Request.URL.Path, "/channel/", "", 1), nil
}
