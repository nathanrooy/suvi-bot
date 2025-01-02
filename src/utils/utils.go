package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

const JPEGQuality int = 80

type Post struct {
	ImgBuf      bytes.Buffer
	Description string
}

func mostRecent(url1 string, url2 string) string {
	t1, _ := time.Parse("20060102T150405Z", strings.Split(url1, "_")[4][1:])
	t2, _ := time.Parse("20060102T150405Z", strings.Split(url2, "_")[4][1:])
	if t1.After(t2) {
		return url1
	} else {
		return url2
	}
}

func getLatestImageURL() string {
	resp, err := http.Get("https://services.swpc.noaa.gov/images/animations/suvi/primary/171/")
	if err != nil {
		log.Println("> could not reach suvi endpoint!")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("> Error reading response body:", err)
	}

	var latest string
	re := regexp.MustCompile(`href="([^"]+)"`)
	for _, url := range re.FindAllStringSubmatch(string(body), -1) {
		if strings.Contains(url[1], "or_suvi") {
			if latest == "" {
				latest = url[1]
			} else {
				latest = mostRecent(url[1], latest)
			}
		}
	}

	return "https://services.swpc.noaa.gov/images/animations/suvi/primary/171/" + latest
}

func treatImage(buf *bytes.Buffer, save bool) *bytes.Buffer {
	reader := bytes.NewReader(buf.Bytes())
	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println("Error decoding image:", err)
	}

	// crop
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	img = imaging.Crop(img, image.Rect(0, 0, w, h-50))
	img = imaging.Paste(
		imaging.New(w, h, color.Black),
		img, image.Pt(0, 0),
	)

	// misc
	img = imaging.AdjustBrightness(img, 4)
	img = imaging.AdjustContrast(img, 2.5)
	img = imaging.AdjustSaturation(img, 10)
	img = imaging.Sharpen(img, 1)

	// dev only
	if save {
		err = imaging.Save(img, "image.jpg", imaging.JPEGQuality(JPEGQuality))
		if err != nil {
			fmt.Println("Error saving image:", err)
		}
	}

	buf = new(bytes.Buffer)
	err = imaging.Encode(buf, img, imaging.JPEG, imaging.JPEGQuality(JPEGQuality))
	if err != nil {
		log.Println("> Error encoding image:", err)
	}
	return buf
}

func CreatePost() Post {
	url := getLatestImageURL()
	log.Println("> Latest image url:", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Println("> failed to download latest image")
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Println("> Error reading image data into buffer:", err)
	}

	var post Post
	_, err = io.Copy(&post.ImgBuf, treatImage(buf, false))
	if err != nil {
		log.Println("> failed to import image:", err)
	}

	d := strings.Split(url, "_")[4][1:]
	post.Description = d[:4] + "-" + d[4:6] + "-" + d[6:8]
	post.Description += " " + d[9:11] + ":" + d[11:13] + ":" + d[13:15] + "Z"
	return post
}
