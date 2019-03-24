package structutils

// Tokenize split struct in array of {Name1, Value1, Name2, Value2..}
func Tokenize(s interface{}) []interface{} {
	iter, err := NewIterator(s, []string{})
	if err != nil {
		return nil
	}
	var is []interface{}

	for {
		f := iter.Next()
		if f == nil {
			break
		}
		is = append(is, f.Name)
		is = append(is, f.Value)
	}
	return is
}
