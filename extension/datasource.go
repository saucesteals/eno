package extension

import (
	"context"
	"errors"
	"net/http"
)

type DataSource struct {
	MDXId         string `json:"mdxId"`
	MDXUrlId      string `json:"mdxUrlId"`
	Name          string `json:"name"`
	MerchantUrl   string `json:"merchantUrl"`
	Color         string `json:"color"`
	ColorContrast string `json:"colorContrast"`
	ImageUrl      string `json:"imageUrl"`
}

func (a *Extension) DataSourceSearch(ctx context.Context, url string) (DataSource, error) {
	type Payload struct {
		URL string `json:"url"`
	}

	payload := Payload{
		URL: url,
	}

	var datasource DataSource
	req, err := a.newWibRequest(ctx, http.MethodPost, "wib/dataSourceSearch", payload)
	if err != nil {
		return datasource, err
	}

	if err := a.do(req, &datasource); err != nil {
		return datasource, err
	}

	if datasource.MDXId == "" {
		return datasource, errors.New("no datasource found")
	}

	return datasource, nil
}
