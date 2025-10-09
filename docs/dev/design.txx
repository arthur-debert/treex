treex Design

This document describes treex's design principles and architecture.It's not an api guide but a high level overview of how and why treex works like this.

1. Origins

    treex was born out of the will to document file trees, as those are incredible useful in the markdwon readme, but lost when working with the project int the actual codebase, in the shell.

    Not only a part, but not co-located meant those were bound do get our of sync. The idea behind treex is that you have a codebase you don't know, $threex will show you a high lever overview of where things are / what they do. The hability to sync (remove inexistent paths) and export as markdown lower the friction even more, and make the vision of code doc with code, shell acessible but easy to pull out a reality.

    With usage, threex evolved to allow more powerful filtering an querying control. With time the codebase became hard to mange, as everyhting was too interwined to the infofile system. As time went on, it became obvious that these were orthogonal: a terminal tree viewer that was easy to extend with rich querying capabilities and a file systme layout that could provide annotations to that tree.

    