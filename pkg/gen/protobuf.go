package gen

import (
	"bufio"
	"io"
	"os"
	"path"
)

func (t *table) GenerateProtobuf(packageName string, writer io.Writer) error {
	buf := bufio.NewWriter(writer)
	buf.WriteString("syntax = \"proto3\";\n\n")
	buf.WriteString("package " + packageName + ";\n\n")
	buf.WriteString(t.protoimports.ProtoString())
	buf.WriteString("message " + CamelCase(t.name) + " {\n")
	for _, column := range t.columns {
		buf.WriteString("\t" + column.GenerateProtobuf())
		buf.WriteString("\n")
	}
	buf.WriteString("}\n")
	buf.Flush()
	return nil
}

func (t *table) GenerateProtobufToDir(packageName, schemaDir string) error {
	f, err := os.OpenFile(path.Join(schemaDir, t.name+".proto"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.GenerateProtobuf(packageName, f)
}
