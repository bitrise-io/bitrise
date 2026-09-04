## upload — copy a file or directory from the user's machine into the session

When the user asks to bring one of their local files into this session — "upload
my cert from ~/Downloads", "send my local config over", "add this file from my
machine" — call upload. The file lives on the user's machine (you can't see it
here); this copies it from there into this session.

```bash
curl -sS -X POST "$URL/upload" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"localPath": "~/Downloads/cert.p12"}'
```

- `localPath` (required): the path **on the user's machine** to copy. The user
  has to name it (you can't list their local files). A relative path is taken
  relative to the project directory they launched from; `~` means their home.
- `remoteFolder` (optional): the destination folder **here in the session**. Add
  it **only** when the user names where the file should go, e.g.
  `-d '{"localPath": "…", "remoteFolder": "/abs/folder/in/this/session"}'`. If you
  omit it the file goes to a temporary folder in the session — deliberately not
  the working/repo directory, so an uploaded file can't be committed by accident.
  Do not default to the current directory.

On success the response is `{"uploaded": true, "remoteFolder": "…"}`. **Always
tell the user the returned `remoteFolder`** — that is where the file actually
landed, and unless they named a destination it is a temp folder, not the
directory you're working in. If the local file doesn't exist the response
explains that — relay it so they can correct the path.
