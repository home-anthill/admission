package validators

import (
	"admission/api"
	"admission/models"
	"math"

	"github.com/go-playground/validator/v10"
)

// RegisterValidations registers API-specific custom validations.
func RegisterValidations(validate *validator.Validate) {
	validate.RegisterStructValidation(validateSpecReq, api.SpecReq{})
}

func validateSpecReq(sl validator.StructLevel) {
	spec := sl.Current().Interface().(api.SpecReq)

	hasMin := spec.Min != nil
	hasMax := spec.Max != nil
	hasStep := spec.Step != nil
	hasList := spec.List != nil

	switch spec.Format {
	case models.List:
		if hasMin {
			sl.ReportError(spec.Min, "min", "Min", "excluded_with_format_list", "")
		}
		if hasMax {
			sl.ReportError(spec.Max, "max", "Max", "excluded_with_format_list", "")
		}
		if hasStep {
			sl.ReportError(spec.Step, "step", "Step", "excluded_with_format_list", "")
		}
		if !hasList || len(spec.List) == 0 {
			sl.ReportError(spec.List, "list", "List", "required_with_format_list", "")
			return
		}

		seen := make(map[int]struct{}, len(spec.List))
		for _, item := range spec.List {
			if item.Value == nil {
				continue
			}

			value := *item.Value
			if _, ok := seen[value]; ok {
				sl.ReportError(item.Value, "value", "Value", "unique", "")
			}
			seen[value] = struct{}{}
		}

	case models.Bool:
		if hasMin {
			sl.ReportError(spec.Min, "min", "Min", "excluded_with_format_bool", "")
		}
		if hasMax {
			sl.ReportError(spec.Max, "max", "Max", "excluded_with_format_bool", "")
		}
		if hasStep {
			sl.ReportError(spec.Step, "step", "Step", "excluded_with_format_bool", "")
		}
		if hasList {
			sl.ReportError(spec.List, "list", "List", "excluded_with_format_bool", "")
		}

	case models.Int, models.Float:
		if !hasMin {
			sl.ReportError(spec.Min, "min", "Min", "required_with_numeric_format", "")
		}
		if !hasMax {
			sl.ReportError(spec.Max, "max", "Max", "required_with_numeric_format", "")
		}
		if !hasStep {
			sl.ReportError(spec.Step, "step", "Step", "required_with_numeric_format", "")
		}
		if hasList {
			sl.ReportError(spec.List, "list", "List", "excluded_with_numeric_format", "")
		}

		if hasMin && hasMax && hasStep {
			validateRange(sl, *spec.Min, *spec.Max, *spec.Step)
		}
	}
}

func validateRange(sl validator.StructLevel, min float64, max float64, step float64) {
	if !isFinite(min) {
		sl.ReportError(min, "min", "Min", "finite", "")
	}
	if !isFinite(max) {
		sl.ReportError(max, "max", "Max", "finite", "")
	}
	if !isFinite(step) {
		sl.ReportError(step, "step", "Step", "finite", "")
	}
	if !isFinite(min) || !isFinite(max) || !isFinite(step) {
		return
	}

	if min >= max {
		sl.ReportError(max, "max", "Max", "gtfield", "min")
		return
	}

	if step <= 0 {
		sl.ReportError(step, "step", "Step", "gt", "0")
		return
	}

	steps := (max - min) / step
	if !isInteger(steps) {
		sl.ReportError(step, "step", "Step", "divides_range", "")
	}
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func isInteger(v float64) bool {
	const epsilon = 1e-9
	return math.Abs(v-math.Round(v)) < epsilon
}
