package mapper_test

import (
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

type (
	TestDTO struct {
		ID        uuid.UUID
		Price     []PriceDTO
		Default   DefaultDTO
		UpdatedAt *time.Time
	}

	DefaultDTO struct {
		Currency string
	}

	PriceDTO struct {
		Price float64
	}

	TestDomain struct {
		ID        uuid.UUID
		UpdatedAt *time.Time
		Price     []model.JSONB
		Default   model.JSONB
	}
)

func Test_MappingDtoToDomain(t *testing.T) {
	id, _ := uuid.FromString("23a1d325-8969-44d6-83be-06b074be58af")
	prices := []model.JSONB{
		model.JSONB{
			"Price": 123.421,
		},
		model.JSONB{
			"Price": 666.666,
		},
	}
	def := model.JSONB{
		"Currency": "USD",
	}

	updatedAt, _ := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")
	updatedAtDto, _ := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")

	dto := TestDTO{
		ID: id,
		Price: []PriceDTO{
			PriceDTO{
				Price: 123.421,
			},
			PriceDTO{
				Price: 666.666,
			},
		},
		UpdatedAt: &updatedAtDto,
		Default: DefaultDTO{
			Currency: "USD",
		},
	}
	domain := TestDomain{}
	err := mapper.Map(dto, &domain)

	assert.Nil(t, err, "Error: %v", err)
	assert.Equal(t, id, domain.ID, "ID not equal")
	assert.Equal(t, prices, domain.Price, "Prices not equal")
	assert.Equal(t, def, domain.Default, "Default not equal")
	assert.Equal(t, &updatedAt, domain.UpdatedAt, "UpdatedAt not equal")

}

func Test_MappingDomainToDto(t *testing.T) {
	id, _ := uuid.FromString("23a1d325-8969-44d6-83be-06b074be58af")
	updatedAt, _ := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")
	updatedAtDomain, _ := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")
	prices := []PriceDTO{
		PriceDTO{
			Price: 123.421,
		},
		PriceDTO{
			Price: 666.666,
		},
	}
	def := DefaultDTO{
		Currency: "USD",
	}

	domain := TestDomain{
		ID: id,
		Default: model.JSONB{
			"Currency": "USD",
		},
		UpdatedAt: &updatedAtDomain,
		Price: []model.JSONB{
			model.JSONB{
				"Price": 123.421,
			},
			model.JSONB{
				"Price": 666.666,
			},
		},
	}

	dto := TestDTO{}
	err := mapper.Map(domain, &dto)

	assert.Nil(t, err, "Error: %v", err)
	assert.Equal(t, id, domain.ID, "ID not equal")
	assert.Equal(t, prices, dto.Price, "Prices not equal")
	assert.Equal(t, def, dto.Default, "Default not equal")
	assert.Equal(t, &updatedAt, dto.UpdatedAt, "UpdatedAt not equal")
}

func Test_ArrayMapping(t *testing.T) {
	prices := []model.JSONB{
		model.JSONB{
			"Price": 123.421,
		},
		model.JSONB{
			"Price": 666.666,
		},
	}
	pricesDto := []PriceDTO{
		PriceDTO{
			Price: 123.421,
		},
		PriceDTO{
			Price: 666.666,
		},
	}

	var dto []PriceDTO
	err := mapper.Map(prices, &dto)

	assert.Nil(t, err, "Error: %v", err)
	assert.Equal(t, pricesDto, dto, "Prices not equal")
}
