package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/h2non/gock"
	"github.com/nbio/st"
	"go.uber.org/zap"
)

// test the makerequest function with the gock library
func TestMakeRequest(t *testing.T) {
	// create mock panda client
	client := &PandaClient{
		Logger:      zap.NewNop().Sugar(),
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		HTTPClient:  &http.Client{},
		DBConnector: &dbtypes.Queries{},
		Run:         0,
		Ctx:         t.Context(),
	}
	gock.InterceptClient(client.HTTPClient)

	gock.New("https://api.pandascore.io").
		MatchHeader("Accept", "application/json").
		MatchHeader("Authorization", "Bearer fakesecret").
		Get("/videogames").
		MatchParams(map[string]string{
			"per_page":   "100",
			"otherparam": "hasvalue",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": "some data",
		})

	resp, err := client.MakeRequest([]string{"videogames"}, map[string]string{
		"otherparam": "hasvalue",
	})
	if err != nil {
		t.Errorf("Error making request: %v", err)
	}
	st.Expect(t, resp.StatusCode, 200)
}

func TestMakeRequestError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// create mock panda client
	client := &PandaClient{
		Logger: zap.NewNop().Sugar(),
		// url.Parse fails when baseurl is not properly formatted
		BaseURL:     "#$%^&*($#$%%^(",
		Pandasecret: "fakesecret",
		HTTPClient:  &http.Client{},
		DBConnector: &dbtypes.Queries{},
		Run:         0,
		Ctx:         ctx,
	}
	// url.Parse fails when baseurl is not properly formatted
	_, err := client.MakeRequest([]string{"videogames"}, map[string]string{
		"otherparam": "hasvalue",
	})
	st.Reject(t, err, nil)
	//malformed URL would fail httpNewRequestWithContext
	client.BaseURL = "https://api.panda score.io"
	_, err = client.MakeRequest([]string{"videogames"}, map[string]string{
		"otherparam": "hasvalue",
	})
	st.Reject(t, err, nil)

	// context failing would fail client.HTTPClient.Do
	cancel()
	_, err = client.MakeRequest([]string{"videogames"}, map[string]string{
		"otherparam": "hasvalue",
	})
	st.Reject(t, err, nil)
}
