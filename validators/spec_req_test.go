package validators

import (
	"admission/api"
	"admission/models"
	"math"
	"math/rand"
	"regexp"
	"testing"

	"github.com/go-playground/validator/v10"
)

var alphanumPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func TestSpecReqValidation(t *testing.T) {
	tests := []struct {
		name    string
		spec    api.SpecReq
		wantErr bool
	}{
		{
			name: "valid list",
			spec: api.SpecReq{
				Format: models.List,
				List: []api.SpecListItemReq{
					listItem(0, "dry"),
					listItem(1, "heat"),
				},
			},
		},
		{
			name: "valid list with non-contiguous values",
			spec: api.SpecReq{
				Format: models.List,
				List: []api.SpecListItemReq{
					listItem(0, "dry"),
					listItem(10, "heat"),
					listItem(42, "cool"),
				},
			},
		},
		{
			name: "valid list with negative values",
			spec: api.SpecReq{
				Format: models.List,
				List: []api.SpecListItemReq{
					listItem(-2, "low"),
					listItem(-1, "mid"),
					listItem(0, "high"),
				},
			},
		},
		{
			name: "valid list with max item count",
			spec: api.SpecReq{
				Format: models.List,
				List:   listItems(20),
			},
		},
		{
			name: "list rejects range fields",
			spec: api.SpecReq{
				Format: models.List,
				Min:    new(0.0),
				List:   []api.SpecListItemReq{listItem(0, "dry")},
			},
			wantErr: true,
		},
		{
			name:    "list requires list field",
			spec:    api.SpecReq{Format: models.List},
			wantErr: true,
		},
		{
			name: "list rejects empty list",
			spec: api.SpecReq{
				Format: models.List,
				List:   []api.SpecListItemReq{},
			},
			wantErr: true,
		},
		{
			name: "list rejects nil item value",
			spec: api.SpecReq{
				Format: models.List,
				List:   []api.SpecListItemReq{{Text: "dry"}},
			},
			wantErr: true,
		},
		{
			name: "list rejects empty item text",
			spec: api.SpecReq{
				Format: models.List,
				List:   []api.SpecListItemReq{listItem(0, "")},
			},
			wantErr: true,
		},
		{
			name: "list rejects duplicate negative values",
			spec: api.SpecReq{
				Format: models.List,
				List: []api.SpecListItemReq{
					listItem(-1, "low"),
					listItem(-1, "mid"),
				},
			},
			wantErr: true,
		},
		{
			name: "list rejects more than max item count",
			spec: api.SpecReq{
				Format: models.List,
				List:   listItems(21),
			},
			wantErr: true,
		},
		{
			name:    "valid bool",
			spec:    api.SpecReq{Format: models.Bool},
			wantErr: false,
		},
		{
			name:    "bool rejects min",
			spec:    api.SpecReq{Format: models.Bool, Min: new(0.0)},
			wantErr: true,
		},
		{
			name:    "bool rejects max",
			spec:    api.SpecReq{Format: models.Bool, Max: new(1.0)},
			wantErr: true,
		},
		{
			name:    "bool rejects step",
			spec:    api.SpecReq{Format: models.Bool, Step: new(1.0)},
			wantErr: true,
		},
		{
			name: "bool rejects list",
			spec: api.SpecReq{
				Format: models.Bool,
				List:   []api.SpecListItemReq{listItem(0, "off")},
			},
			wantErr: true,
		},
		{
			name: "valid int",
			spec: api.SpecReq{
				Format: models.Int,
				Min:    new(0.0),
				Max:    new(100.0),
				Step:   new(1.0),
			},
		},
		{
			name: "valid float",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(-40.0),
				Max:    new(200.0),
				Step:   new(0.01),
			},
		},
		{
			name: "valid float with negative range",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(-10.0),
				Max:    new(-1.0),
				Step:   new(3.0),
			},
		},
		{
			name: "valid float with epsilon-sized division remainder",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Max:    new(0.3),
				Step:   new(0.1),
			},
		},
		{
			name: "numeric requires min",
			spec: api.SpecReq{
				Format: models.Float,
				Max:    new(200.0),
				Step:   new(0.01),
			},
			wantErr: true,
		},
		{
			name: "numeric requires max",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Step:   new(0.01),
			},
			wantErr: true,
		},
		{
			name: "numeric requires step",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Max:    new(200.0),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects list",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Max:    new(10.0),
				Step:   new(1.0),
				List:   []api.SpecListItemReq{listItem(0, "dry")},
			},
			wantErr: true,
		},
		{
			name: "numeric rejects non-dividing step",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Max:    new(10.0),
				Step:   new(3.0),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects equal min and max",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(10.0),
				Max:    new(10.0),
				Step:   new(1.0),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects min greater than max",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(11.0),
				Max:    new(10.0),
				Step:   new(1.0),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects zero step",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Max:    new(10.0),
				Step:   new(0.0),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects negative step",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Max:    new(10.0),
				Step:   new(-1.0),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects infinite max",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Max:    new(math.Inf(1)),
				Step:   new(1.0),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects infinite min",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(math.Inf(-1)),
				Max:    new(10.0),
				Step:   new(1.0),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects infinite step",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(0.0),
				Max:    new(10.0),
				Step:   new(math.Inf(1)),
			},
			wantErr: true,
		},
		{
			name: "numeric rejects nan",
			spec: api.SpecReq{
				Format: models.Float,
				Min:    new(math.NaN()),
				Max:    new(10.0),
				Step:   new(1.0),
			},
			wantErr: true,
		},
		{
			name: "list rejects duplicate values",
			spec: api.SpecReq{
				Format: models.List,
				List: []api.SpecListItemReq{
					listItem(0, "dry"),
					listItem(0, "heat"),
				},
			},
			wantErr: true,
		},
		{
			name: "list rejects non-alphanumeric text",
			spec: api.SpecReq{
				Format: models.List,
				List:   []api.SpecListItemReq{listItem(0, "too hot")},
			},
			wantErr: true,
		},
		{
			name:    "format is required",
			spec:    api.SpecReq{},
			wantErr: true,
		},
		{
			name:    "unknown format is rejected",
			spec:    api.SpecReq{Format: models.SpecFormat("string")},
			wantErr: true,
		},
	}

	validate := newSpecReqValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSpecReqValidation(t, validate, tt.spec, tt.wantErr)
		})
	}
}

func TestSpecReqValidationGeneratedPayloads(t *testing.T) {
	validate := newSpecReqValidator()
	rng := rand.New(rand.NewSource(42))

	for range 1000 {
		spec := randomSpecReq(rng)
		wantErr := !isExpectedValidSpecReq(spec)
		assertSpecReqValidation(t, validate, spec, wantErr)
	}
}

func newSpecReqValidator() *validator.Validate {
	validate := validator.New()
	RegisterValidations(validate)
	return validate
}

func assertSpecReqValidation(t *testing.T, validate *validator.Validate, spec api.SpecReq, wantErr bool) {
	t.Helper()

	err := validate.Struct(spec)
	if wantErr && err == nil {
		t.Fatalf("expected validation error for spec %#v", spec)
	}
	if !wantErr && err != nil {
		t.Fatalf("expected no validation error for spec %#v, got %v", spec, err)
	}
}

func isExpectedValidSpecReq(spec api.SpecReq) bool {
	if spec.Format != models.Bool && spec.Format != models.Int && spec.Format != models.Float && spec.Format != models.List {
		return false
	}

	if spec.List != nil {
		if len(spec.List) > 20 {
			return false
		}
		for _, item := range spec.List {
			if item.Value == nil || !alphanumPattern.MatchString(item.Text) {
				return false
			}
		}
	}

	switch spec.Format {
	case models.List:
		if spec.Min != nil || spec.Max != nil || spec.Step != nil || len(spec.List) == 0 {
			return false
		}

		seen := make(map[int]struct{}, len(spec.List))
		for _, item := range spec.List {
			value := *item.Value
			if _, ok := seen[value]; ok {
				return false
			}
			seen[value] = struct{}{}
		}
		return true

	case models.Bool:
		return spec.Min == nil && spec.Max == nil && spec.Step == nil && spec.List == nil

	case models.Int, models.Float:
		if spec.Min == nil || spec.Max == nil || spec.Step == nil || spec.List != nil {
			return false
		}
		if !isFinite(*spec.Min) || !isFinite(*spec.Max) || !isFinite(*spec.Step) {
			return false
		}
		if *spec.Min >= *spec.Max || *spec.Step <= 0 {
			return false
		}
		return isInteger((*spec.Max - *spec.Min) / *spec.Step)
	}

	return false
}

func randomSpecReq(rng *rand.Rand) api.SpecReq {
	formats := []models.SpecFormat{
		"",
		models.Bool,
		models.Int,
		models.Float,
		models.List,
		models.SpecFormat("string"),
	}

	spec := api.SpecReq{
		Format: formats[rng.Intn(len(formats))],
	}

	if rng.Intn(2) == 0 {
		spec.Min = randomFloatPtr(rng)
	}
	if rng.Intn(2) == 0 {
		spec.Max = randomFloatPtr(rng)
	}
	if rng.Intn(2) == 0 {
		spec.Step = randomFloatPtr(rng)
	}
	if rng.Intn(2) == 0 {
		spec.List = randomListItems(rng)
	}

	return spec
}

func randomFloatPtr(rng *rand.Rand) *float64 {
	values := []float64{
		-10,
		-1,
		0,
		0.1,
		1,
		3,
		10,
		math.Inf(-1),
		math.Inf(1),
		math.NaN(),
	}
	return new(values[rng.Intn(len(values))])
}

func randomListItems(rng *rand.Rand) []api.SpecListItemReq {
	items := make([]api.SpecListItemReq, rng.Intn(23))
	texts := []string{"", "dry", "too hot", "Heat2", "cool", "low"}

	for i := range items {
		if rng.Intn(4) != 0 {
			items[i].Value = new(rng.Intn(7) - 3)
		}
		items[i].Text = texts[rng.Intn(len(texts))]
	}

	return items
}

func listItems(count int) []api.SpecListItemReq {
	items := make([]api.SpecListItemReq, count)
	for i := range count {
		items[i] = api.SpecListItemReq{
			Value: new(i),
			Text:  "item" + string(rune('A'+i)),
		}
	}
	return items
}

func listItem(value int, text string) api.SpecListItemReq {
	return api.SpecListItemReq{
		Value: new(value),
		Text:  text,
	}
}
