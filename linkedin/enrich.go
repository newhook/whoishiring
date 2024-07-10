package linkedin

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"os"
)

var apiKey = os.Getenv("PROXYCURL_API_KEY")

func Person(ctx context.Context, linkedInProfileUrl string) (*PersonEndpointResponse, error) {
	if false {
		body, err := os.ReadFile("response.json")
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Unmarshal response into ApiResponse struct
		var apiResponse PersonEndpointResponse
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			return nil, errors.WithStack(err)
		}

		return &apiResponse, nil
	}

	baseUrl := "https://nubela.co/proxycurl/api/v2/linkedin"

	// Prepare URL with query parameters
	queryParams := url.Values{}
	//queryParams.Add("twitter_profile_url", "https://x.com/johnrmarty/")
	//queryParams.Add("facebook_profile_url", "https://facebook.com/johnrmarty/")
	queryParams.Add("linkedin_profile_url", linkedInProfileUrl)
	//queryParams.Add("extra", "include")
	//queryParams.Add("github_profile_id", "linkedInProfileUrl")
	//queryParams.Add("facebook_profile_id", "include")
	//queryParams.Add("twitter_profile_id", "include")
	//queryParams.Add("personal_contact_number", "include")
	//queryParams.Add("personal_email", "include")
	//queryParams.Add("inferred_salary", "include")
	//queryParams.Add("skills", "include")
	//queryParams.Add("use_cache", "if-present")
	//queryParams.Add("fallback_to_cache", "on-error")

	// Construct request
	req, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+apiKey)

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response from the embedding API: " + resp.Status)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = os.WriteFile("response.json", body, 0644)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Unmarshal response into ApiResponse struct
	var apiResponse PersonEndpointResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, errors.WithStack(err)
	}

	return &apiResponse, nil
}
