package storage

import (
	"context"
	"fmt"
	"reflect"
	"time"

	tarantool "github.com/tarantool/go-tarantool/v2"
)

func SendMessageTT(ctx context.Context, userID, to, text string) error {
	dialogID := GetDialogId(userID, to)
	_, err := tt.Do(tarantool.NewCallRequest("send_message").
		Args([]interface{}{userID, to, dialogID, text}),
	).Get()

	if err != nil {
		fmt.Println("TT: Error while send_message()")
		return err
	}
	return nil
}

func DialogListTT(ctx context.Context, userID, to string) ([]SendRequest, error) {
	dialogID := GetDialogId(userID, to)

	resp, err := tt.Do(tarantool.NewCallRequest("get_dialog").
		Args([]interface{}{dialogID}),
	).Get()
	if err != nil {
		fmt.Println("TT: Error while get_dialog()")
		return nil, err
	}
	var dialog []SendRequest
	baseData := resp.Data[0]
	var baseOut []interface{}

	rv := reflect.ValueOf(baseData)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			baseOut = append(baseOut, rv.Index(i).Interface())
		}
	}

	for _, data := range baseOut {
		var out []interface{}
		rv := reflect.ValueOf(data)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				out = append(out, rv.Index(i).Interface())
			}
		}
		dialog = append(dialog,
			SendRequest{AuthorID: out[1].(string),
				RecepientID: out[2].(string),
				DialogID:    out[3].(string),
				CreatedAt:   time.Unix(int64(out[4].(uint32)), 0),
				Text:        out[5].(string)})
	}

	return dialog, nil
}
