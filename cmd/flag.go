package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

type Flags struct {
	namespace     string
	allNamespaces bool
	label         string
	fieldSelector string
}

func (f Flags) Namespace() string {
	if f.allNamespaces {
		return ""
	}
	return f.namespace
}

func (f Flags) Label() string {
	return f.label
}

func (f Flags) FieldSelector(args []string) string {
	for _, v := range args {
		fieldSelectors := strings.Split(f.fieldSelector, ",")
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("metadata.name=%s", v))
		f.fieldSelector = strings.Join(fieldSelectors, ",")
	}
	return f.fieldSelector
}

func InitFlags(cmd *cobra.Command, flags *Flags) {
	cmd.Flags().StringVarP(
		&flags.namespace,
		"namespace",
		"n",
		"default",
		"kubernetes namespace",
	)
	cmd.Flags().BoolVarP(
		&flags.allNamespaces,
		"all-namespaces",
		"A",
		false,
		"all kubernetes namespaces",
	)
	cmd.Flags().StringVarP(
		&flags.label,
		"label",
		"l",
		"",
		"kubernetes label",
	)
	cmd.Flags().StringVarP(
		&flags.fieldSelector,
		"field-selector",
		"",
		"",
		"kubernetes field selector",
	)
}
