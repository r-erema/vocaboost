package spacedrepetition

import (
	"context"
	"fmt"

	"github.com/jomei/notionapi"
	"github.com/r-erema/vocaboost/internal/domain"
)

type Notion struct {
	client     *notionapi.Client
	databaseID notionapi.DatabaseID
}

func NewNotion(client *notionapi.Client, databaseID notionapi.DatabaseID) *Notion {
	return &Notion{client: client, databaseID: databaseID}
}

func (n Notion) UploadWords(ctx context.Context, words []*domain.Word) error {
	for _, word := range words {
		if _, err := n.client.Page.Create(ctx, n.buildPageCreateRequest(word)); err != nil {
			return fmt.Errorf("the Notion page for word `%s` creation error: %w", word.Word(), err)
		}
	}

	return nil
}

func (n Notion) buildPageCreateRequest(word *domain.Word) *notionapi.PageCreateRequest {
	blocks := make([]notionapi.Block, 0)

	for _, imageURL := range word.ImageURLs() {
		blocks = append(blocks, imageBlock(imageURL))
	}

	for _, definition := range word.Definitions() {
		blocks = append(blocks, definitionBlock(definition))
	}

	for _, example := range word.Examples() {
		blocks = append(blocks, exampleBlock(example))
	}

	return &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: n.databaseID,
			PageID:     "",
		},
		Properties: map[string]notionapi.Property{
			"Name": notionapi.TitleProperty{
				ID:   "",
				Type: "",
				Title: []notionapi.RichText{
					{Text: notionapi.Text{Content: word.Word(), Link: nil}},
				},
			},
		},
		Children: blocks,
		Icon:     nil,
		Cover:    nil,
	}
}

func imageBlock(imageURL string) notionapi.ImageBlock {
	return notionapi.ImageBlock{
		BasicBlock: notionapi.BasicBlock{
			Object:         notionapi.ObjectTypeBlock,
			Type:           notionapi.BlockTypeImage,
			ID:             "",
			CreatedTime:    nil,
			LastEditedTime: nil,
			CreatedBy:      nil,
			LastEditedBy:   nil,
			HasChildren:    false,
			Archived:       false,
		},
		Image: notionapi.Image{
			Type: notionapi.FileTypeExternal,
			External: &notionapi.FileObject{
				URL:        imageURL,
				ExpiryTime: nil,
			},
			Caption: nil,
			File:    nil,
		},
	}
}

func definitionBlock(definition string) notionapi.BulletedListItemBlock {
	return notionapi.BulletedListItemBlock{
		BasicBlock: notionapi.BasicBlock{
			Object:         notionapi.ObjectTypeBlock,
			Type:           notionapi.BlockTypeBulletedListItem,
			ID:             "",
			CreatedTime:    nil,
			LastEditedTime: nil,
			CreatedBy:      nil,
			LastEditedBy:   nil,
			HasChildren:    false,
			Archived:       false,
		},
		BulletedListItem: notionapi.ListItem{
			RichText: []notionapi.RichText{
				{
					Text: notionapi.Text{Content: definition, Link: nil},
				},
			},
			Children: nil,
			Color:    "",
		},
	}
}

func exampleBlock(example string) notionapi.ParagraphBlock {
	return notionapi.ParagraphBlock{
		BasicBlock: notionapi.BasicBlock{
			Object:         notionapi.ObjectTypeBlock,
			Type:           notionapi.BlockTypeParagraph,
			ID:             "",
			CreatedTime:    nil,
			LastEditedTime: nil,
			CreatedBy:      nil,
			LastEditedBy:   nil,
			HasChildren:    false,
			Archived:       false,
		},
		Paragraph: notionapi.Paragraph{
			RichText: []notionapi.RichText{
				{
					Text: notionapi.Text{Content: example, Link: nil},
					Annotations: &notionapi.Annotations{
						Italic:        true,
						Bold:          false,
						Strikethrough: false,
						Underline:     false,
						Code:          false,
						Color:         "",
					},
				},
			},
			Children: nil,
			Color:    "",
		},
	}
}
