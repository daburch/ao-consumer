package neo4j

import (
	"fmt"
	"strings"

	"github.com/daburch/ao_tools/ao_consumer/pkg/models"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func ProcessMatch(transaction neo4j.Transaction, match models.CrystalLeagueMatch) {

	transaction.Run(fmt.Sprintf("MERGE (n:CrystalMatch { matchId: '%s' }) SET n.level=%d, n.time='%s', n.winner=%d, n.team1Tickets=%d, n.team2Tickets=%d",
		match.MatchID, match.CrystalLeagueLevel, match.MatchTime, match.Winner, match.Team1.Tickets, match.Team2.Tickets), nil)

	for _, p := range match.Team1.Players {
		ProcessPlayer(transaction, match.MatchID, p, 1)
	}

	for _, p := range match.Team2.Players {
		ProcessPlayer(transaction, match.MatchID, p, 2)
	}
}

func ProcessPlayer(transaction neo4j.Transaction, matchId string, player models.Player, team int) {
	transaction.Run(fmt.Sprintf("MERGE (a:Player { name: '%s' }) SET a.displayName = '%s'", strings.ToLower(player.Name), player.Name), nil)
	transaction.Run(fmt.Sprintf("MATCH (a:Player), (b:CrystalMatch) WHERE a.name = '%s' AND b.matchId = '%s' MERGE (a)-[r:COMPETED_IN]->(b) SET r.team = %d, r.kills = %d, r.deaths = %d, r.healing = %d",
		strings.ToLower(player.Name), matchId, team, player.Kills, player.Deaths, player.Healing), nil)
}
