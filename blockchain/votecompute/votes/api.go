package votes

import (
	"github.com/pkt-cash/pktd/blockchain"
	"github.com/pkt-cash/pktd/blockchain/votecompute/db"
	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/database"
	"github.com/pkt-cash/pktd/pktlog/log"
)

func DisconnectBlock(
	tx database.Tx,
	block *btcutil.Block,
	spent []blockchain.SpentTxOut,
) er.R {
	votes, err := parseVotes(block, spent)
	if err != nil {
		return err
	}
	for _, v := range votes {
		if err := db.DeleteVote(tx, int32(v.VoteCastInBlock), v.VoterPkScript); err != nil {
			return err
		}
	}
	return nil
}

func ConnectBlock(dbTx database.Tx, block *btcutil.Block, stxo []blockchain.SpentTxOut) er.R {
	votes, err := parseVotes(block, stxo)
	if err != nil {
		log.Errorf("Unable to parse votes from block number [%d]: [%s]", block.Height(), err)
		return err
	}
	for _, v := range votes {
		if err := db.PutVote(
			dbTx,
			int32(v.VoteCastInBlock),
			v.VoterPkScript,
			v.VoterIsWillingCandidate,
			v.VoterPkScript,
		); err != nil {
			log.Errorf("Unable to store votes from block number [%d]: [%s]", block.Height(), err)
			return err
		}
	}
	return nil
}
