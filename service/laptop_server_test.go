package service

import (
	"context"
	"testing"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/JeongWoo-Seo/pcBook/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLaptopServer(t *testing.T) {
	t.Parallel()

	laptopNoId := util.NewLaptop()
	laptopNoId.Id = ""

	laptopInvalidId := util.NewLaptop()
	laptopInvalidId.Id = "invalid"

	laptopDuplicateId := util.NewLaptop()
	storeDuplicateId := NewInMemoryLaptopStore()
	err := storeDuplicateId.Save(laptopDuplicateId)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		laptop      *pb.Laptop
		laptopstore LaptopStore
		code        codes.Code
	}{
		{
			name:        "ok",
			laptop:      util.NewLaptop(),
			laptopstore: NewInMemoryLaptopStore(),
			code:        codes.OK,
		},
		{
			name:        "no_id",
			laptop:      laptopNoId,
			laptopstore: NewInMemoryLaptopStore(),
			code:        codes.OK,
		},
		{
			name:        "invalid_id",
			laptop:      laptopInvalidId,
			laptopstore: NewInMemoryLaptopStore(),
			code:        codes.InvalidArgument,
		},
		{
			name:        "duplicate_id",
			laptop:      laptopDuplicateId,
			laptopstore: storeDuplicateId,
			code:        codes.AlreadyExists,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}

			server := NewLaptopServer(tc.laptopstore, nil)
			res, err := server.CreateLaptop(context.Background(), req)
			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res.Id)
				if len(tc.laptop.Id) > 0 {
					require.Equal(t, tc.laptop.Id, res.Id)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.code, st.Code())
			}

		})
	}

}
