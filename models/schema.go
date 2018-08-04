package models

import (
	"log"
	"strconv"

	graphql "github.com/neelance/graphql-go"
)

var Schema = `
`

type Resolver struct{}

type postResolver struct {
	p *Post
}

func (p *postResolver) Id() graphql.ID {
	return graphql.ID(strconv.FormatInt(p.p.Id, 10))
}

func (p *postResolver) Title() string {
	return p.p.Title
}

func (p *postResolver) Content() string {
	return p.p.Content
}

// TODO: Make actual html
func (p *postResolver) Html() string {
	return string(p.p.Html())
}

// TODO: Make actual html
func (p *postResolver) SummaryHtml() string {
	return SummarizeText(p.p.Content)
}

func (p *postResolver) Datetime() graphql.Time {
	return graphql.Time{Time: p.p.Datetime}
}

func (p *postResolver) Readtime() int32 {
	return p.p.ReadTime()
}

func (p *postResolver) Created() graphql.Time {
	return graphql.Time{Time: p.p.Created}
}

func (p *postResolver) Modified() graphql.Time {
	return graphql.Time{Time: p.p.Modified}
}

func (p *postResolver) Draft() bool {
	return p.p.Draft
}

func (r *Resolver) Post(args struct{ ID graphql.ID }) (*postResolver, error) {
	id, err := strconv.ParseInt(string(args.ID), 10, 64)
	if err != nil {
		log.Printf("Got error trying to convert id: %+v: %+v", args.ID, err)
		return nil, err
	}

	post, err := GetPost(id)
	if err != nil {
		log.Printf("Got error trying to get post %+v: %+v", args.ID, err)
		return nil, err
	}
	return &postResolver{post}, nil
}

func (r *Resolver) Posts() ([]*postResolver, error) {
	posts, err := AllPosts()
	if err != nil {
		log.Printf("Got error trying to get posts: %+v", err)
		return nil, err
	}

	l := make([]*postResolver, len(posts))
	for i, p := range posts {
		l[i] = &postResolver{p}
	}
	return l, nil
}
