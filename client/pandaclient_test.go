package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/h2non/gock"
	"github.com/nbio/st"
	"go.uber.org/zap/zaptest"
)

// test the makerequest function with the gock library
func TestMakeRequest(t *testing.T) {
	// create mock panda client
	client := &PandaClient{
		Logger:      zaptest.NewLogger(t).Sugar(),
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

// tests the cases where makerequest function fails
func TestMakeRequestError(t *testing.T) {
	expected := (*http.Response)(nil)
	ctx, cancel := context.WithCancel(context.Background())
	// create mock panda client
	client := &PandaClient{
		Logger: zaptest.NewLogger(t).Sugar(),
		// url.Parse fails when baseurl is not properly formatted
		BaseURL:     "#$%^&*($#$%%^(",
		Pandasecret: "fakesecret",
		HTTPClient:  &http.Client{},
		DBConnector: &dbtypes.Queries{},
		Run:         0,
		Ctx:         nil,
	}
	// url.Parse fails when baseurl is not properly formatted
	res, err := client.MakeRequest([]string{"videogames"}, map[string]string{
		"otherparam": "hasvalue",
	})
	st.Expect(t, res, expected)
	st.Reject(t, err, nil)
	//nil context
	client.BaseURL = "https://api.pandascore.io"
	res, err = client.MakeRequest([]string{"videogames"}, map[string]string{
		"otherparam": "hasvalue",
	})
	st.Expect(t, res, expected)
	st.Reject(t, err, nil)
	client.Ctx = ctx
	// context failing would fail client.HTTPClient.Do
	cancel()
	res, err = client.MakeRequest([]string{"videogames"}, map[string]string{
		"otherparam": "hasvalue",
	})
	st.Expect(t, res, expected)
	st.Reject(t, err, nil)
}
