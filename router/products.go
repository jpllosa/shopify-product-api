package router

import (
	"fmt"
	"io"
	"net/http"
	"shopify-product-api/config"
	"shopify-product-api/marshaller"
	"shopify-product-api/service"
	"strings"

	"github.com/go-chi/chi/v5"
)

type GqlQuery struct {
	Query string `json:"query"`
}

type GqlResponse struct {
	Data       Data       `json:"data,omitempty"`
	Extensions Extensions `json:"extensions,omitempty"`
}

type Data struct {
	Products Products `json:"products,omitempty"`
}

type Products struct {
	Edges []Edge `json:"edges,omitempty"`
}

type Edge struct {
	Node Node `json:"node,omitempty"`
}

type Node struct {
	Id          string   `json:"id,omitempty"`
	Title       string   `json:"title,omitempty"`
	Handle      string   `json:"handle,omitempty"`
	Vendor      string   `json:"vendor,omitempty"`
	ProductType string   `json:"productType,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type Extensions struct {
	Cost Cost `json:"cost,omitempty"`
}

type Cost struct {
	RequestedQueryCost int            `json:"requestedQueryCost,omitempty"`
	ActualQueryCost    int            `json:"actualQueryCost,omitempty"`
	ThrottleStatus     ThrottleStatus `json:"throttleStatus,omitempty"`
}

type ThrottleStatus struct {
	MaximumAvailable   float32 `json:"maximumAvailable,omitempty"`
	CurrentlyAvailable int     `json:"currentlyAvailable,omitempty"`
	RestoreRate        float32 `json:"restoreRate,omitempty"`
}

const contentType string = "Content-Type"
const applicationJson string = "application/json"

func GetProducts(config config.Config) chi.Router {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// 	# Get the ID and title of the three most recently added products
		// curl -X POST   https://{store_name}.myshopify.com/admin/api/2023-10/graphql.json \
		//   -H 'Content-Type: application/json' \
		//   -H 'X-Shopify-Access-Token: {access_token}' \
		//   -d '{
		//   "query": "{
		//     products(first: 3) {
		//       edges {
		//         node {
		//           id
		//           title
		//         }
		//       }
		//     }
		//   }"
		// }'
		productsGql := fmt.Sprintf(`{
			products(first: 3, reverse: false) {
			  edges {
				node {
				  id
				  title
				  handle
				  vendor
				  productType
				  tags
				}
			  }
			}
		  }`)

		query := GqlQuery{
			Query: productsGql,
		}

		q := marshaller.Marshal(query)
		body := strings.NewReader(string(q))

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPost, config.Shopify.Endpoint, body)
		if err != nil {
			panic(err)
		}

		req.Header.Add(contentType, applicationJson)
		req.Header.Add("X-Shopify-Access-Token", config.Shopify.AccessToken)

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println("Response status:", resp.Status)

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		gqlResp := marshaller.Unmarshal[GqlResponse](responseBody)

		jsonBytes := marshaller.Marshal(gqlResp.Data.Products.Edges)

		w.Header().Set(contentType, applicationJson)
		w.WriteHeader(200)
		w.Write(jsonBytes)
	})

	return router
}

func GetCachedProducts(config config.Config, products []service.Product) chi.Router {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {

		jsonBytes := marshaller.Marshal(products)

		w.Header().Set(contentType, applicationJson)
		w.WriteHeader(200)
		w.Write(jsonBytes)
	})

	return router
}
