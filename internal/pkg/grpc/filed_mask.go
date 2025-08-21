package grpc

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type CompiledFieldMask struct {
	Paths []FieldPath
}

type FieldPath struct {
	FieldName string
	Nested    *CompiledFieldMask // 嵌套字段
}

func CompiledFiledMask(mask *fieldmaskpb.FieldMask) (*CompiledFieldMask, error) {
	compiled := &CompiledFieldMask{}

	for _, path := range mask.Paths {
		parts := strings.Split(path, ".")
		current := compiled

		for i, part := range parts {
			var found *FieldPath
			for _, p := range current.Paths {
				if p.FieldName == part {
					found = &p
					break
				}
			}

			if found == nil {
				newPath := FieldPath{FieldName: part}
				current.Paths = append(current.Paths, newPath)
				found = &newPath
			}

			// 最后一部分，不需要嵌套。
			if i < len(parts)-1 {
				if found.Nested == nil {
					found.Nested = &CompiledFieldMask{}
				}
				current = found.Nested
			}
		}
	}
	return compiled, nil
}

func ApplyFieldMaskToRepeated(src, dst protoreflect.Message, mask *fieldmaskpb.FieldMask) error {
	if mask == nil {
		return nil
	}

	srcRef := src.Interface().ProtoReflect()
	dstRef := dst.Interface().ProtoReflect()

	for _, path := range mask.GetPaths() {
		if err := applyFieldMaskToRepeated(srcRef, dstRef, path); err != nil {
			return err
		}
	}

	return nil
}

func applyFieldMaskToRepeated(src, dst protoreflect.Message, path string) error {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil
	}

	fieldName := protoreflect.Name(parts[0])
	field := src.Descriptor().Fields().ByName(fieldName)
	if field == nil {
		return fmt.Errorf("cannot find field [ %s ]", parts[0])
	}

	if field.Cardinality() == protoreflect.Repeated {
		srcList := src.Get(field).List()
		dstList := dst.Get(field).List()

		for i := range srcList.Len() {
			if i >= dstList.Len() {
				newElem := dstList.NewElement()
				dstList.Append(newElem)
			}

			srcElem := srcList.Get(i)
			dstElem := dstList.Get(i)

			if field.Message() != nil && len(parts) > 1 {
				nestedPath := strings.Join(parts[1:], ".")
				nestedMask := &fieldmaskpb.FieldMask{
					Paths: []string{nestedPath},
				}

				err := ApplyFieldMaskToRepeated(
					srcElem.Message().Interface().ProtoReflect(),
					dstElem.Message().Interface().ProtoReflect(),
					nestedMask,
				)
				if err != nil {
					return err
				}
			} else if len(parts) == 1 {
				dstList.Set(i, srcElem)
			}

		}
	} else if field.Message() != nil && len(parts) > 1 {
		// 处理嵌套字段
		nestedPath := strings.Join(parts[1:], ".")
		nestedMask := &fieldmaskpb.FieldMask{
			Paths: []string{nestedPath},
		}

		srcMessage := src.Get(field).Message().Interface()
		dstMessage := dst.Mutable(field).Message().Interface()

		return ApplyFieldMaskToRepeated(
			srcMessage.ProtoReflect(),
			dstMessage.ProtoReflect(),
			nestedMask,
		)
	} else {
		dst.Set(field, src.Get(field))
	}

	return nil
}
