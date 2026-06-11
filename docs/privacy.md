# Privacy

TraceDeck collects typed endpoint metadata only. It does not collect passwords,
keystrokes, browser cookies, auth tokens, private messages, camera, microphone,
or covert screenshots.

Sensitive capabilities are deny-only in typed policy. Media file name/path and
browser video metadata require explicit policy enablement. Browser activity is
domain/category based by default, not full URL based.

The process collector stores app/process names and executable path hashes. It
does not persist raw executable paths.
