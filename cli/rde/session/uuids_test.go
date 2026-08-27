package session

// uuidSession is a UUID-shaped session arg. Real RDE session IDs are UUIDs,
// so this short-circuits ResolveSessionID's name lookup.
const uuidSession = "99999999-8888-7777-6666-555555555555"
