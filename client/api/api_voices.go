/*
 * ElevenLabs API Documentation
 *
 * This is the documentation for the ElevenLabs API. You can use this API to use our service programmatically, this is done by using your xi-api-key. <br/> You can view your xi-api-key using the 'Profile' tab on https://beta.elevenlabs.io. Our API is experimental so all endpoints are subject to change.
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/antihax/optional"
)

// Linger please
var (
	_ context.Context
)

type VoicesApiService service

/*
VoicesApiService Add Voice
Add a new voice to your collection of voices in VoiceLab.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param name
 * @param files
 * @param description
 * @param labels
 * @param optional nil or *VoicesApiAddVoiceV1VoicesAddPostOpts - Optional Parameters:
     * @param "XiApiKey" (optional.String) -  Your API key. This is required by most endpoints to access our API programatically. You can view your xi-api-key using the &#x27;Profile&#x27; tab on the website.
@return AddVoiceResponseModel
*/

type VoicesApiAddVoiceV1VoicesAddPostOpts struct {
	XiApiKey optional.String
}

func (a *VoicesApiService) AddVoiceV1VoicesAddPost(ctx context.Context, name string, files []*os.File, description string, labels string, localVarOptionals *VoicesApiAddVoiceV1VoicesAddPostOpts) (AddVoiceResponseModel, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Post")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue AddVoiceResponseModel
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/voices/add"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"multipart/form-data"}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XiApiKey.IsSet() {
		localVarHeaderParams["xi-api-key"] = parameterToString(localVarOptionals.XiApiKey.Value(), "")
	}
	localVarFormParams.Add("name", parameterToString(name, ""))
	localVarFormParams.Add("files", parameterToString(files, "multi"))
	localVarFormParams.Add("description", parameterToString(description, ""))
	localVarFormParams.Add("labels", parameterToString(labels, ""))
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v AddVoiceResponseModel
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 422 {
			var v HttpValidationError
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
VoicesApiService Delete Voice
Deletes a voice by its ID.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param voiceId Voice ID to be used, you can use https://api.elevenlabs.io/v1/voices to list all the available voices.
 * @param optional nil or *VoicesApiDeleteVoiceV1VoicesVoiceIdDeleteOpts - Optional Parameters:
     * @param "XiApiKey" (optional.String) -  Your API key. This is required by most endpoints to access our API programatically. You can view your xi-api-key using the &#x27;Profile&#x27; tab on the website.
@return Object
*/

type VoicesApiDeleteVoiceV1VoicesVoiceIdDeleteOpts struct {
	XiApiKey optional.String
}

func (a *VoicesApiService) DeleteVoiceV1VoicesVoiceIdDelete(ctx context.Context, voiceId string, localVarOptionals *VoicesApiDeleteVoiceV1VoicesVoiceIdDeleteOpts) (Object, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Delete")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Object
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}"
	localVarPath = strings.Replace(localVarPath, "{"+"voice_id"+"}", fmt.Sprintf("%v", voiceId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XiApiKey.IsSet() {
		localVarHeaderParams["xi-api-key"] = parameterToString(localVarOptionals.XiApiKey.Value(), "")
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v Object
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 422 {
			var v HttpValidationError
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
VoicesApiService Edit Voice Settings
Edit your settings for a specific voice. \&quot;similarity_boost\&quot; corresponds to\&quot;Clarity + Similarity Enhancement\&quot; in the web app and \&quot;stability\&quot; corresponds to \&quot;Stability\&quot; slider in the web app.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param body
 * @param voiceId Voice ID to be used, you can use https://api.elevenlabs.io/v1/voices to list all the available voices.
 * @param optional nil or *VoicesApiEditVoiceSettingsV1VoicesVoiceIdSettingsEditPostOpts - Optional Parameters:
     * @param "XiApiKey" (optional.String) -  Your API key. This is required by most endpoints to access our API programatically. You can view your xi-api-key using the &#x27;Profile&#x27; tab on the website.
@return Object
*/

type VoicesApiEditVoiceSettingsV1VoicesVoiceIdSettingsEditPostOpts struct {
	XiApiKey optional.String
}

func (a *VoicesApiService) EditVoiceSettingsV1VoicesVoiceIdSettingsEditPost(ctx context.Context, body Settings, voiceId string, localVarOptionals *VoicesApiEditVoiceSettingsV1VoicesVoiceIdSettingsEditPostOpts) (Object, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Post")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Object
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/settings/edit"
	localVarPath = strings.Replace(localVarPath, "{"+"voice_id"+"}", fmt.Sprintf("%v", voiceId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XiApiKey.IsSet() {
		localVarHeaderParams["xi-api-key"] = parameterToString(localVarOptionals.XiApiKey.Value(), "")
	}
	// body params
	localVarPostBody = &body
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v Object
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 422 {
			var v HttpValidationError
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
VoicesApiService Edit Voice
Edit a voice created by you.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param name
 * @param files
 * @param description
 * @param labels
 * @param voiceId Voice ID to be used, you can use https://api.elevenlabs.io/v1/voices to list all the available voices.
 * @param optional nil or *VoicesApiEditVoiceV1VoicesVoiceIdEditPostOpts - Optional Parameters:
     * @param "XiApiKey" (optional.String) -  Your API key. This is required by most endpoints to access our API programatically. You can view your xi-api-key using the &#x27;Profile&#x27; tab on the website.
@return Object
*/

type VoicesApiEditVoiceV1VoicesVoiceIdEditPostOpts struct {
	XiApiKey optional.String
}

func (a *VoicesApiService) EditVoiceV1VoicesVoiceIdEditPost(ctx context.Context, name string, files []*os.File, description string, labels string, voiceId string, localVarOptionals *VoicesApiEditVoiceV1VoicesVoiceIdEditPostOpts) (Object, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Post")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Object
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/edit"
	localVarPath = strings.Replace(localVarPath, "{"+"voice_id"+"}", fmt.Sprintf("%v", voiceId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"multipart/form-data"}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XiApiKey.IsSet() {
		localVarHeaderParams["xi-api-key"] = parameterToString(localVarOptionals.XiApiKey.Value(), "")
	}
	localVarFormParams.Add("name", parameterToString(name, ""))
	localVarFormParams.Add("files", parameterToString(files, "multi"))
	localVarFormParams.Add("description", parameterToString(description, ""))
	localVarFormParams.Add("labels", parameterToString(labels, ""))
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v Object
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 422 {
			var v HttpValidationError
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
VoicesApiService Get Default Voice Settings.
Gets the default settings for voices. \&quot;similarity_boost\&quot; corresponds to\&quot;Clarity + Similarity Enhancement\&quot; in the web app and \&quot;stability\&quot; corresponds to \&quot;Stability\&quot; slider in the web app.
  - @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().

@return VoiceSettingsResponseModel
*/
func (a *VoicesApiService) GetDefaultVoiceSettingsV1VoicesSettingsDefaultGet(ctx context.Context) (VoiceSettingsResponseModel, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue VoiceSettingsResponseModel
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/voices/settings/default"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v VoiceSettingsResponseModel
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
VoicesApiService Get Voice Settings
Returns the settings for a specific voice. \&quot;similarity_boost\&quot; corresponds to\&quot;Clarity + Similarity Enhancement\&quot; in the web app and \&quot;stability\&quot; corresponds to \&quot;Stability\&quot; slider in the web app.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param voiceId Voice ID to be used, you can use https://api.elevenlabs.io/v1/voices to list all the available voices.
 * @param optional nil or *VoicesApiGetVoiceSettingsV1VoicesVoiceIdSettingsGetOpts - Optional Parameters:
     * @param "XiApiKey" (optional.String) -  Your API key. This is required by most endpoints to access our API programatically. You can view your xi-api-key using the &#x27;Profile&#x27; tab on the website.
@return VoiceSettingsResponseModel
*/

type VoicesApiGetVoiceSettingsV1VoicesVoiceIdSettingsGetOpts struct {
	XiApiKey optional.String
}

func (a *VoicesApiService) GetVoiceSettingsV1VoicesVoiceIdSettingsGet(ctx context.Context, voiceId string, localVarOptionals *VoicesApiGetVoiceSettingsV1VoicesVoiceIdSettingsGetOpts) (VoiceSettingsResponseModel, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue VoiceSettingsResponseModel
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/settings"
	localVarPath = strings.Replace(localVarPath, "{"+"voice_id"+"}", fmt.Sprintf("%v", voiceId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XiApiKey.IsSet() {
		localVarHeaderParams["xi-api-key"] = parameterToString(localVarOptionals.XiApiKey.Value(), "")
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v VoiceSettingsResponseModel
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 422 {
			var v HttpValidationError
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
VoicesApiService Get Voice
Returns metadata about a specific voice.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param voiceId Voice ID to be used, you can use https://api.elevenlabs.io/v1/voices to list all the available voices.
 * @param optional nil or *VoicesApiGetVoiceV1VoicesVoiceIdGetOpts - Optional Parameters:
     * @param "WithSettings" (optional.Bool) -  If set will return settings information corresponding to the voice, requires authorization.
     * @param "XiApiKey" (optional.String) -  Your API key. This is required by most endpoints to access our API programatically. You can view your xi-api-key using the &#x27;Profile&#x27; tab on the website.
@return VoiceResponseModel
*/

type VoicesApiGetVoiceV1VoicesVoiceIdGetOpts struct {
	WithSettings optional.Bool
	XiApiKey     optional.String
}

func (a *VoicesApiService) GetVoiceV1VoicesVoiceIdGet(ctx context.Context, voiceId string, localVarOptionals *VoicesApiGetVoiceV1VoicesVoiceIdGetOpts) (VoiceResponseModel, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue VoiceResponseModel
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}"
	localVarPath = strings.Replace(localVarPath, "{"+"voice_id"+"}", fmt.Sprintf("%v", voiceId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if localVarOptionals != nil && localVarOptionals.WithSettings.IsSet() {
		localVarQueryParams.Add("with_settings", parameterToString(localVarOptionals.WithSettings.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XiApiKey.IsSet() {
		localVarHeaderParams["xi-api-key"] = parameterToString(localVarOptionals.XiApiKey.Value(), "")
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v VoiceResponseModel
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 422 {
			var v HttpValidationError
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
VoicesApiService Get Voices
Gets a list of all available voices for a user.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param optional nil or *VoicesApiGetVoicesV1VoicesGetOpts - Optional Parameters:
     * @param "XiApiKey" (optional.String) -  Your API key. This is required by most endpoints to access our API programatically. You can view your xi-api-key using the &#x27;Profile&#x27; tab on the website.
@return GetVoicesResponseModel
*/

type VoicesApiGetVoicesV1VoicesGetOpts struct {
	XiApiKey optional.String
}

func (a *VoicesApiService) GetVoicesV1VoicesGet(ctx context.Context, localVarOptionals *VoicesApiGetVoicesV1VoicesGetOpts) (GetVoicesResponseModel, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue GetVoicesResponseModel
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/voices"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XiApiKey.IsSet() {
		localVarHeaderParams["xi-api-key"] = parameterToString(localVarOptionals.XiApiKey.Value(), "")
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v GetVoicesResponseModel
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 422 {
			var v HttpValidationError
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}