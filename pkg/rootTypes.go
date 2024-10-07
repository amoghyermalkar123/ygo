package pkg

type TextRef struct {
	branch *Branch
}

func NewTextRef() TextRef {
	return TextRef{
		branch: &Branch{},
	}
}
