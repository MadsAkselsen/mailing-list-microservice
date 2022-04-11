package grpcapi

import (
	"database/sql"
	"mailinglist/mdb"
	pb "mailinglist/proto"
	"time"
)

type MailServer struct {
	pb.UnimplementedMailingListServiceServer
	db *sql.DB
}

func pbEntryToMdbEntry(pbEntry *pb.EmailEntry) mdb.EmailEntry {
	t := time.Unix(pbEntry.ConfirmedAt, 0)
	return mdb.EmailEntry{
		Id: pbEntry.Id,
		Email: pbEntry.Email,
		ConfirmedAt: &t,
		OptOut: pbEntry.OptOut,
	}
}

func mdbEntryToPbEntry(mdbEntry *mdb.EmailEntry) pb.EmailEntry {
	return pb.EmailEntry{
		Id: mdbEntry.Id,
		Email: mdbEntry.Email,
		ConfirmedAt: mdbEntry.ConfirmedAt.Unix(),
		OptOut: mdbEntry.OptOut,
	}
}