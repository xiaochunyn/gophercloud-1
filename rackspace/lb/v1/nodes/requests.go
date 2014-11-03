package nodes

import (
	"errors"
	"fmt"

	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace/lb/v1"
)

func List(client *gophercloud.ServiceClient, loadBalancerID int, limit *int) pagination.Pager {
	url := rootURL(client, loadBalancerID)
	if limit != nil {
		url += fmt.Sprintf("?limit=%d", limit)
	}

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return NodePage{pagination.SinglePageBase(r)}
	})
}

type CreateOptsBuilder interface {
	ToNodeCreateMap() (map[string]interface{}, error)
}

type CreateOpts []CreateOpt

type CreateOpt struct {
	// Required
	Address   string
	Port      int
	Condition Condition
	Type      Type
}

func (opts CreateOpts) ToNodeCreateMap() (map[string]interface{}, error) {
	type nodeMap map[string]interface{}
	nodes := []nodeMap{}

	for k, v := range opts {
		if v.Address == "" {
			return nodeMap{}, fmt.Errorf("ID is a required attribute, none provided for %d CreateOpt element", k)
		}

		node := make(map[string]interface{})
		node["address"] = v.Address

		if v.Port > 0 {
			node["port"] = v.Port
		}
		if v.Condition != "" {
			node["condition"] = v.Condition
		}
		if v.Type != "" {
			node["type"] = v.Type
		}

		nodes = append(nodes, node)
	}

	return nodeMap{"nodes": nodes}, nil
}

func Create(client *gophercloud.ServiceClient, loadBalancerID int, opts CreateOptsBuilder) CreateResult {
	var res CreateResult

	reqBody, err := opts.ToNodeCreateMap()
	if err != nil {
		res.Err = err
		return res
	}

	resp, err := perigee.Request("POST", rootURL(client, loadBalancerID), perigee.Options{
		MoreHeaders: client.AuthenticatedHeaders(),
		ReqBody:     &reqBody,
		Results:     &res.Body,
		OkCodes:     []int{200},
	})
	if err != nil {
		res.Err = err
		return res
	}

	pr, err := pagination.PageResultFrom(resp.HttpResponse)
	if err != nil {
		res.Err = err
		return res
	}

	return CreateResult{pagination.SinglePageBase(pr)}
}

func BulkDelete(c *gophercloud.ServiceClient, loadBalancerID int, nodeIDs []int) DeleteResult {
	var res DeleteResult

	if len(nodeIDs) > 10 || len(nodeIDs) == 0 {
		res.Err = errors.New("You must provide a minimum of 1 and a maximum of 10 node IDs")
		return res
	}

	url := rootURL(c, loadBalancerID)
	url += v1.IDSliceToQueryString("id", nodeIDs)

	_, res.Err = perigee.Request("DELETE", url, perigee.Options{
		MoreHeaders: c.AuthenticatedHeaders(),
		OkCodes:     []int{202},
	})

	return res
}

func Get(c *gophercloud.ServiceClient, lbID, nodeID int) GetResult {
	var res GetResult

	_, res.Err = perigee.Request("GET", resourceURL(c, lbID, nodeID), perigee.Options{
		MoreHeaders: c.AuthenticatedHeaders(),
		Results:     &res.Body,
		OkCodes:     []int{200},
	})

	return res
}
