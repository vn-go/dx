package from

import "fmt"

func (f *fromClauseInfo) String() string {
	source := fmt.Sprintf("%s %s JOIN %s on %s", f.left, f.joinType, f.right, f.on)
	if f.next != nil {
		f = f.next
		return f.nextString(source)
	} else {
		return source
	}
	// if f.left != "" {
	// 	return fmt.Sprintf("%s %s JOIN %s on %s", f.left, f.joinType, f.right, f.on)
	// } else {
	// 	return fmt.Sprintf("%s JOIN %s on %s", f.joinType, f.right, f.on)
	// }

}

func (f *fromClauseInfo) nextString(left string) string {
	source := fmt.Sprintf("%s %s JOIN %s on %s", left, f.joinType, f.right, f.on)
	if f.next != nil {
		f = f.next
		return f.nextString(source)
	} else {
		return source
	}
}
