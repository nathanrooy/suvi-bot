package bsky

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"suvi/src/utils"
	"time"
)

const PDS_URL string = "https://bsky.social"

type Session struct {
	DID       string `json:"did"`
	AccessJWT string `json:"accessJwt"`
}

type EmbedResponse struct {
	Blob Embed `json:"blob"`
}

type Embed struct {
	Type string `json:"$type"`
	Ref  struct {
		Link string `json:"$link"`
	} `json:"ref"`
	Mimetype string `json:"mimeType"`
	Size     int64  `json:"size"`
}

func _login() Session {
	var session Session
	postPayload := []byte(`{
		"identifier":"` + os.Getenv("BSKY_USER") + `",
		"password":"` + os.Getenv("BSKY_PSWD") + `"
	}`)

	req, err := http.NewRequest(
		"POST",
		PDS_URL+"/xrpc/com.atproto.server.createSession",
		bytes.NewBuffer(postPayload),
	)
	if err != nil {
		log.Println("> Error creating request:", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("> Error sending request:", err)
	}
	defer resp.Body.Close()

	log.Println("> BSKY login response code:", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("> Error reading response body:", err)
	}

	err = json.Unmarshal(body, &session)
	if err != nil {
		log.Println("> Error marshaling json response:", err)
	}

	return session
}

func _uploadImage(s Session, p utils.Post) Embed {
	var mimetype string = "image/png"

	req, err := http.NewRequest(
		"POST",
		PDS_URL+"/xrpc/com.atproto.repo.uploadBlob",
		&p.ImgBuf,
	)
	if err != nil {
		log.Println("> Error creating request:", err)
	}

	req.Header.Set("Content-Type", mimetype)
	req.Header.Set("Authorization", "Bearer "+s.AccessJWT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("> Error sending request:", err)
	}
	defer resp.Body.Close()

	log.Println("> BSKY image upload response code:", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("> Error reading response body:", err)
	}

	var embedResp EmbedResponse
	err = json.Unmarshal(body, &embedResp)
	if err != nil {
		log.Println("> Error marshalling json response:", err)
	}

	return embedResp.Blob

}

func _post(s Session, p utils.Post) {

	embed := _uploadImage(s, p)

	images := map[string]interface{}{
		"alt":   p.Description,
		"image": embed,
	}

	post := map[string]interface{}{
		"repo":       s.DID,
		"collection": "app.bsky.feed.post",
		"record": map[string]interface{}{
			"embed": map[string]interface{}{
				"$type":  "app.bsky.embed.images",
				"images": [1]map[string]interface{}{images},
			},
			"$type":     "app.bsky.feed.post",
			"text":      p.Description + " #NASA #NOAA #GOES16 #Space",
			"createdAt": time.Now().UTC().Format("2006-01-02T15:04:05Z"),
			"facets": []map[string]interface{}{
				{
					"index": map[string]int{
						"byteStart": 21,
						"byteEnd":   26,
					},
					"features": []map[string]string{
						{
							"$type": "app.bsky.richtext.facet#tag",
							"tag":   "nasa",
						},
					},
				},
				{
					"index": map[string]int{
						"byteStart": 27,
						"byteEnd":   32,
					},
					"features": []map[string]string{
						{
							"$type": "app.bsky.richtext.facet#tag",
							"tag":   "noaa",
						},
					},
				},
				{
					"index": map[string]int{
						"byteStart": 33,
						"byteEnd":   40,
					},
					"features": []map[string]string{
						{
							"$type": "app.bsky.richtext.facet#tag",
							"tag":   "goes16",
						},
					},
				},
				{
					"index": map[string]int{
						"byteStart": 41,
						"byteEnd":   47,
					},
					"features": []map[string]string{
						{
							"$type": "app.bsky.richtext.facet#tag",
							"tag":   "space",
						},
					},
				},
			},
		},
	}

	postJSON, err := json.Marshal(post)
	if err != nil {
		log.Println("> Error marshalling JSONL", err)
	}

	req, err := http.NewRequest(
		"POST",
		PDS_URL+"/xrpc/com.atproto.repo.createRecord",
		bytes.NewBuffer(postJSON),
	)
	if err != nil {
		log.Println("> Error creating request:", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.AccessJWT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("> Error sending request:", err)
	}
	defer resp.Body.Close()

	log.Println("> BSKY create post response code:", resp.StatusCode)

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println("> Error reading response body:", err)
	}
}

func _purge(s Session) {

}

func Run(p utils.Post) {

	s := _login()
	_post(s, p)

}
