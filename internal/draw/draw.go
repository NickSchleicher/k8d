package draw

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/NickSchleicher/k8d/internal/k8s"
)

const format = `## **********************************************************
## Configuration
## **********************************************************
# label: %name%<br>%notes%
# style: whiteSpace=wrap;html=1;rounded=1;fillColor=#ffffff;strokeColor=#000000;
# namespace: csvimport-
# connect: {"from": "refs", "to": "id", "style": "fontSize=11;"}
# connect: {"from": "ingress", "to": "id", "invert": true, "style": "dashed=1;fontSize=11;"}
# connect: {"from": "egress", "to": "id", "style": "dashed=1;fontSize=11;"}
# width: auto
# height: auto
# padding: 5
# ignore: id,refs,ingress,egress
# link: url
# nodespacing: 60
# levelspacing: 60
# edgespacing: 40
# layout: auto
## **********************************************************
## CSV Data
## **********************************************************
id,name,notes,refs,ingress,egress
`

type draw struct {
	cluster  *k8s.Cluster
	fileName *string
	rows     []row
}

type row struct {
	name    string
	notes   string
	refs    []string
	ingress []string
	egress  []string
}

// Output creates the draw.io text file
func Output(c *k8s.Cluster, f *string) error {
	d := &draw{
		cluster:  c,
		fileName: f,
	}

	d.addNamespaces()

	err := d.writeOut()
	if err != nil {
		return err
	}

	return nil
}

func (d *draw) addNamespaces() {
	for _, ns := range d.cluster.Namespaces {
		d.rows = append(d.rows, row{
			name: ns.Name,
		})

		d.addPods(d.getIndex(), ns)
	}
}

func (d *draw) addPods(index int, ns *k8s.Namespace) {
	for _, p := range ns.Pods {
		d.rows = append(d.rows, row{
			name:  p.Name,
			notes: ns.Name,
		})

		d.rows[index].refs = append(d.rows[index].refs, strconv.Itoa(d.getIndex()))

		podIndex := d.getIndex()
		d.rows[podIndex].ingress = d.addRules(p.Ingress)

		// d.addRules(k, p, p.Egress)
	}
}

func (d *draw) addRules(rules []k8s.Rule) []string {
	var refs []string

	for _, r := range rules {
		if r.Everywhere {
			d.rows = append(d.rows, row{
				name: "Everywhere",
			})

			refs = append(refs, strconv.Itoa(d.getIndex()))
		} else {
			if r.IPBlock != nil {
				d.rows = append(d.rows, row{
					name:  r.IPBlock.CIDR,
					notes: strings.Join(r.IPBlock.Except, "\n"),
				})

				refs = append(refs, strconv.Itoa(d.getIndex()))
			}

			for _, p := range r.Pods {
				d.rows = append(d.rows, row{
					name:  p.Name,
					notes: p.Namespace.Name,
				})

				refs = append(refs, strconv.Itoa(d.getIndex()))
			}
		}
	}

	return refs
}

func (d *draw) getIndex() int {
	return len(d.rows) - 1
}

func (d *draw) writeOut() error {
	file, err := os.Create(*d.fileName)
	if err != nil {
		return err
	}

	file.WriteString(format)

	writer := csv.NewWriter(file)
	defer writer.Flush()
	for k, r := range d.rows {
		writer.Write([]string{
			strconv.Itoa(k),
			r.name,
			r.notes,
			strings.Join(r.refs, ","),
			strings.Join(r.ingress, ","),
			strings.Join(r.egress, ","),
		})
	}

	return nil
}
