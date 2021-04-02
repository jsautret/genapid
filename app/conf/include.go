// Process the 'include' statement to replace it by the content of
// included file

package conf

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type include struct {
	Include string
}

func processInclude(cfg *yaml.Node) {
	if cfg.Kind == yaml.DocumentNode && len(cfg.Content) == 1 {
		processNode(cfg.Content[0])
	}
}

func processNode(n *yaml.Node) {
	if n.Kind != yaml.SequenceNode {
		log.Fatal().Err(errors.
			New("Toplevel conf & 'pipe' content must be a list")).
			Int("line", n.Line).Msg("")
	}
	l := len(n.Content)
	for i := 0; i < l; i++ {
		p := n.Content[i]
		if p.Kind != yaml.MappingNode {
			log.Fatal().
				Err(errors.New("Predicate must be dict")).
				Int("line", p.Line).Msgf("")
		}
		for j := 0; j < len(p.Content); j += 2 {
			name := ""
			if err := (p.Content[j]).Decode(&name); err != nil {
				log.Fatal().Err(err).
					Int("line", p.Content[j].Line).Msg("")
			}
			if name == "include" {
				nodes := readInclude(n.Content[i])
				// Replace the 'include' by the included predicates
				n.Content = append(n.Content[:i], append(nodes,
					n.Content[i+1:]...)...)
				l += len(nodes) - 1
				i--
			}
			if name == "pipe" {
				processNode(p.Content[j+1])
			}
		}
	}
}

func readInclude(node *yaml.Node) []*yaml.Node {
	inc := include{}
	if err := node.Decode(&inc); err != nil {
		log.Fatal().Err(err).Int("line", node.Line).Msg("Invalid 'include'")
	}
	d := ReadFile(inc.Include)
	if d.Kind != yaml.DocumentNode || len(d.Content) != 1 {
		log.Fatal().
			Err(fmt.Errorf("included file %v is not a valid document",
				inc.Include)).Msg("")
	}
	n := d.Content[0]
	if n.Kind != yaml.SequenceNode {
		log.Fatal().
			Err(fmt.Errorf("included file %v must be a list", inc.Include)).
			Msg("")
	}
	return n.Content
}
