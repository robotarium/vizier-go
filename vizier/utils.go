package vizier

import (
	"errors"
	"fmt"
)

type Link struct {
	_type string
}

type Request struct {
	_type    string
	link     string
	required bool
}

func ParseDescriptor(descriptor map[string]interface{}) (map[string]Link, map[string]Request, error) {

	var endpoint interface{}
	var ok bool
	if endpoint, ok = descriptor["end_point"]; !ok {
		errorMsg := "Node descriptor must contain field endpoint"
		return nil, nil, errors.New(errorMsg)
	}

	var endpointCast string
	if endpointCast, ok = endpoint.(string); !ok {
		errorMsg := fmt.Sprintf("Value for field endpoint (%v) could not be cast to string", endpoint)
		return nil, nil, errors.New(errorMsg)
	}

	parsedLinks, err := parseDescriptorHelper("", endpointCast, descriptor)

	if err != nil {
		return nil, nil, err
	}

	requestsMap := make(map[string]Request)

	if requests, ok := descriptor["requests"]; ok {

		var requestsCast []interface{}
		if requestsCast, ok = requests.([]interface{}); !ok {
			errorMsg := fmt.Sprintf("Node descriptor has field 'request' but could not cast value (%v) to []map[string]interface{}", requests)
			fmt.Println(errorMsg)
			return nil, nil, errors.New(errorMsg)
		}

		for i := range requestsCast {
			var parsedRequest Request
			var err error
			if requestCast, ok := requestsCast[i].(map[string]interface{}); ok {

				if parsedRequest, err = parseRequest(requestCast); err != nil {
					return nil, nil, err
				}

				link := parsedRequest.link
				requestsMap[link] = parsedRequest
			} else {
				errorMsg := "Could not cast request."
				fmt.Println(errorMsg)
				return nil, nil, errors.New(errorMsg)
			}
		}
	}

	return parsedLinks, requestsMap, nil
}

func isSubsetOf(left string, right string) bool {

	lenLeft := len(left)
	lenRight := len(right)

	if lenLeft > lenRight {
		return false
	}

	// Else
	return left == right[:lenLeft]
}

func parseLink(body map[string]interface{}) (link Link, err error) {

	if _type, ok := body["type"]; ok {
		if typeCast, ok := _type.(string); ok {
			link._type = typeCast
		} else {
			errorMsg := fmt.Sprintf("Field type with value (%v) could not be cast to string", _type)
			err = errors.New(errorMsg)
		}

	} else {
		errorMsg := "Base link must contain field type."
		err = errors.New(errorMsg)
	}

	return link, err
}

func parseRequest(body map[string]interface{}) (request Request, err error) {

	if _type, ok := body["type"]; ok {
		if typeCast, ok := _type.(string); ok {
			request._type = typeCast
		} else {
			errorMsg := fmt.Sprintf("Request has type field but value (%v) cannot be cast to string", _type)
			err = errors.New(errorMsg)
		}
	} else {
		errorMsg := "Request must contain field 'type.'"
		err = errors.New(errorMsg)
	}

	if link, ok := body["link"]; ok {
		if linkCast, ok := link.(string); ok {
			request.link = linkCast
		} else {
			errorMsg := fmt.Sprintf("Request has link field but value (%v) cannot be cast to string", link)
			err = errors.New(errorMsg)
		}
	} else {
		errorMsg := "Request must contain field 'link.'."
		err = errors.New(errorMsg)
	}

	if required, ok := body["required"]; ok {
		if requiredCast, ok := required.(bool); ok {
			request.required = requiredCast
		} else {
			errorMsg := fmt.Sprintf("Request has requied field but value (%v) cannot be cast to string", required)
			err = errors.New(errorMsg)
		}
	}

	return
}

func parseDescriptorHelper(path string, link string, body map[string]interface{}) (map[string]Link, error) {

	var links interface{}
	var ok bool
	if links, ok = body["links"]; !ok {
		ret := make(map[string]Link)
		link, err := parseLink(body)

		if err != nil {
			return nil, err
		}

		ret[path] = link
		return ret, nil
	}

	var linksCast map[string]interface{}
	if linksCast, ok = links.(map[string]interface{}); !ok {
		errorMsg := "Links cannot be case to map[string]interface{}"
		fmt.Println(errorMsg)
		return nil, errors.New(errorMsg)
	}

	if len(linksCast) == 0 {
		ret := make(map[string]Link)
		link, err := parseLink(body)

		if err != nil {
			return nil, err
		}

		ret[path] = link
		return ret, nil
	}

	if len(link) == 0 {
		// REturn error
		errorMsg := fmt.Sprintf("Cannot have link of length zero")
		return nil, errors.New(errorMsg)
	}

	// Else
	var pathHere string

	if link[0] == '/' {
		pathHere = path + "/" + link
	} else {
		if isSubsetOf(path, link) {
			pathHere = link
		} else {
			errorMsg := "If not relative, path (%v) must be subset of link (%v)."
			return nil, errors.New(errorMsg)
		}
	}

	parsedLinks := make(map[string]Link)

	for k, v := range linksCast {
		if sublinksCast, ok := v.(map[string]interface{}); ok {
			parsedSubLinks, err := parseDescriptorHelper(pathHere, k, sublinksCast)

			if err != nil {
				return nil, err
			}

			for x, y := range parsedSubLinks {
				parsedLinks[x] = y
			}
		}
	}

	return parsedLinks, nil
}
