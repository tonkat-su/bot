package interactions

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

func (srv *Server) version(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	var (
		commitHash     string
		buildTimestamp string
	)
	buildinfo, available := debug.ReadBuildInfo()
	if !available {
		writeResponse(w, http.StatusOK, "build info not available")
	}
	for _, setting := range buildinfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			commitHash = setting.Value
		case "vcs.time":
			buildTimestamp = setting.Value
		}
	}
	writeResponse(w, http.StatusOK, fmt.Sprintf("froggyfren with commit: %s, timestamp: %s", commitHash, buildTimestamp))
}
