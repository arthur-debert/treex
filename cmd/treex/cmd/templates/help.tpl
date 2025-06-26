{{.Short}}

Usage: 
  $ {{.CommandPath}}{{if .HasAvailableSubCommands}}
  $ {{.CommandPath}} add <path> <description>{{end}}
  
{{range $group := .Groups}}{{if ne $group.ID "main"}}{{.Title}}{{range $cmd := $.Commands}}{{if (and (eq $cmd.GroupID $group.ID) (or $cmd.IsAvailableCommand (eq $cmd.Name "help")))}}
    {{rpad $cmd.Name $.NamePadding }} {{$cmd.Short}}{{end}}{{end}}{{if eq $group.ID "help"}}
    {{rpad "help" $.NamePadding }} Help about any command{{end}}
{{end}}{{end}} 