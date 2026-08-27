package session

// uuidTemplate is a UUID-shaped template arg so ResolveTemplateID
// short-circuits without an extra ListTemplates call.
const uuidTemplate = "11111111-2222-3333-4444-555555555555"

// uuidSession is a UUID-shaped session arg. Real RDE session IDs are UUIDs,
// so this short-circuits ResolveSessionID's name lookup.
const uuidSession = "99999999-8888-7777-6666-555555555555"
