package service

import (
	"fmt"
	"io"
	"net/http"
	"shopify-product-api/config"
	"shopify-product-api/marshaller"
	"strings"
)

func CopyProduct(config config.Config, product Product) error {
	// https://shopify.dev/docs/api/admin-graphql/2024-01/mutations/productCreate
	fmt.Println("++ copy product")

	metafieldStr := ""
	var sb strings.Builder

	for i := 0; i < len(product.Metafields); i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("{")
		sb.WriteString(fmt.Sprintf(`namespace: "%s",`, product.Metafields[i].Namespace))
		sb.WriteString(fmt.Sprintf(`key: "%s",`, product.Metafields[i].Key))
		sb.WriteString(fmt.Sprintf(`value: "%s",`, product.Metafields[i].Value))
		sb.WriteString(fmt.Sprintf(`type: "%s"`, "single_line_text_field")) // assume all metafields are of this type
		sb.WriteString("}")
	}

	if len(product.Metafields) > 0 {
		metafieldStr = fmt.Sprintf(`metafields: [%s]`, sb.String())
	}

	productCreateGql := fmt.Sprintf(`
	mutation { 
		productCreate(
			input: {
				title: "%s",
				productType: "%s",
				vendor: "%s",
				tags: "%s",
				%s
			}
		) {
			product {
				id
			}
		}
	}
	`, product.Title, product.ProductType,
		product.Vendor, strings.Join(product.Tags, ","),
		metafieldStr)

	query := GqlQuery{
		Query: productCreateGql,
	}

	client := &http.Client{}

	responseBody, err := sendCopyRequest(client, query, config)
	if err != nil {
		return err
	}

	gqlResp := marshaller.Unmarshal[GqlResponse](responseBody)

	fmt.Println("New product copied: ", gqlResp.Data.ProductCreate.Product.ID)

	return nil
}

func sendCopyRequest(client *http.Client, query GqlQuery, config config.Config) ([]byte, error) {
	q := marshaller.Marshal(query)
	body := strings.NewReader(string(q))

	req, err := http.NewRequest(http.MethodPost, config.Shopify.TargetEndpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add(contentType, applicationJson)
	req.Header.Add("X-Shopify-Access-Token", config.Shopify.TargetAccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)

	return io.ReadAll(resp.Body)
}
