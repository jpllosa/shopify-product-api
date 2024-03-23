package service

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"shopify-product-api/config"
	"shopify-product-api/marshaller"
	"strings"
	"time"
)

type GqlQuery struct {
	Query string `json:"query"`
}

type GqlResponse struct {
	Data Data `json:"data,omitempty"`
	// Extensions Extensions `json:"extensions,omitempty"`
}

type ProductCreate struct {
	Product Product `json:"product,omitempty"`
}

type Data struct {
	BulkOperationRunQuery BulkOperationRunQuery `json:"bulkOperationRunQuery,omitempty"`
	CurrentBulkOperation  BulkOperation         `json:"currentBulkOperation,omitempty"`
	ProductCreate         ProductCreate         `json:"productCreate,omitempty"`
}

type BulkOperationRunQuery struct {
	BulkOperation BulkOperation `json:"bulkOperation,omitempty"`
	UserErrors    []UserError   `json:"userErrors,omitempty"`
}

type BulkOperation struct {
	CompletedAt string `json:"completedAt,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	ErrorCode   string `json:"errorCode,omitempty"`
	FileSize    string `json:"fileSize,omitempty"`
	ID          string `json:"id,omitempty"`
	ObjectCount string `json:"objectCount,omitempty"`
	Status      string `json:"status,omitempty"`
	URL         string `json:"url,omitempty"`
}

type UserError struct {
	Field   []string `json:"fields,omitempty"`
	Message string   `json:"message,omitempty"`
}

const contentType string = "Content-Type"
const applicationJson string = "application/json"

func BulkQuery(config config.Config) []Product {
	// https://shopify.dev/docs/api/usage/bulk-operations/queries
	fmt.Println("++ bulk query")
	bulkQueryGql := fmt.Sprintf(`
	mutation {
		bulkOperationRunQuery(
			query: """
				{
					products {
						edges {
							node {
								id
								title
								handle
								vendor
								productType
								tags
								metafields {
									edges {
										node {
											namespace
											key
											value
										}
									}
								}
							}
						}
					}
				}
			""") {
			bulkOperation {
				createdAt
				errorCode
				fileSize
				id
				objectCount
				status
				type
				url
			}
			userErrors {
				field
				message
			}
		}
	}
	`)

	query := GqlQuery{
		Query: bulkQueryGql,
	}

	client := &http.Client{}

	responseBody, err := sendRequest(client, query, config)
	if err != nil {
		panic(err)
	}

	gqlResp := marshaller.Unmarshal[GqlResponse](responseBody)

	if gqlResp.Data.BulkOperationRunQuery.BulkOperation.Status == "CREATED" {
		fmt.Println("Created at: ", gqlResp.Data.BulkOperationRunQuery.BulkOperation.CreatedAt)
		currentOperationQueryGql := fmt.Sprintf(`
		query CurrentBulkOperation {
			currentBulkOperation {
				completedAt
				createdAt
				errorCode
				fileSize
				id
				objectCount
				status
				url
			}
		}
		`)

		query = GqlQuery{
			Query: currentOperationQueryGql,
		}

		for {
			time.Sleep(time.Second * 2)

			responseBody, err := sendRequest(client, query, config)
			if err != nil {
				panic(err)
			}

			gqlResp = marshaller.Unmarshal[GqlResponse](responseBody)

			// https://shopify.dev/docs/api/admin-graphql/2023-10/enums/BulkOperationStatus
			if gqlResp.Data.CurrentBulkOperation.Status == "CANCELED" ||
				gqlResp.Data.CurrentBulkOperation.Status == "CANCELING" ||
				gqlResp.Data.CurrentBulkOperation.Status == "EXPIRED" ||
				gqlResp.Data.CurrentBulkOperation.Status == "FAILED" {
				fmt.Println("Status: ", gqlResp.Data.CurrentBulkOperation.CreatedAt)
				break
			}

			if gqlResp.Data.CurrentBulkOperation.Status == "COMPLETED" {
				fmt.Println("URL: ", gqlResp.Data.CurrentBulkOperation.URL)
				productFile, err := downloadFile("products.tmp", gqlResp.Data.CurrentBulkOperation.URL)
				if err != nil {
					break
				}
				return parseProductsFile(productFile)
			}
		}
	}

	return make([]Product, 0)
}

func sendRequest(client *http.Client, query GqlQuery, config config.Config) ([]byte, error) {
	q := marshaller.Marshal(query)
	body := strings.NewReader(string(q))

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

	return io.ReadAll(resp.Body)
}

func downloadFile(filename string, url string) (*os.File, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	file, err := createTemporaryFile(filename)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(file, resp.Body)

	file.Seek(0, io.SeekStart)

	return file, err
}

func createTemporaryFile(filename string) (*os.File, error) {
	file, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func parseProductsFile(file *os.File) []Product {
	defer file.Close()

	var products = make([]Product, 0)
	var metafields = make([]Metafield, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		node := marshaller.Unmarshal[Node](scanner.Bytes())
		if node.ParentID == "" {
			products = append(products, toProduct(node))
		} else {
			metafields = append(metafields, toMetafield(node))
		}
	}

	for i := 0; i < len(products); i++ {
		for _, metafield := range metafields {
			if metafield.ParentID == products[i].ID {
				metafield.ParentID = ""
				products[i].Metafields = append(products[i].Metafields, metafield)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return products
}

func toMetafield(node Node) Metafield {
	return Metafield{
		Namespace: node.Namespace,
		Key:       node.Key,
		Value:     node.Value,
		ParentID:  node.ParentID,
	}
}

func toProduct(node Node) Product {
	return Product{
		ID:          node.ID,
		Title:       node.Title,
		Handle:      node.Handle,
		Vendor:      node.Vendor,
		ProductType: node.ProductType,
		Tags:        node.Tags,
		Metafields:  make([]Metafield, 0),
	}
}
