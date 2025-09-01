E.164:
Is a format for international public telecommunication Phone numbers
ITU-T Recommendation

Canonical format (what your APIs want):
A leading “+” followed by up to 15 digits.
Regex shape: ^\+[1-9]\d{1,14}$

Country calling code (CC): 1–3 digits (India = 91, UK = 44).

National Significant Number (NSN): the rest (which may include an area code/NDC + subscriber number).

Not included: spaces, dashes, parentheses, trunk prefixes (like a leading 0 in many countries), or extensions.

All major SMS providers require E.164 for destinations.

So what we need to do is Normalize whatever user gives into the format of the E.164

Then Validate the number Syntactically possible Reachable type and active

When you send an SMS via a provider (Twilio, Vonage, etc.), you don’t instantly know if it actually reached the handset.
That’s where DLRs (Delivery Receipts) come in.

Lifecycle of an SMS

You send SMS → provider accepts (HTTP 200 OK).

Provider attempts delivery → routes message through carrier networks.

Carrier responds back with final status (delivered / failed).

Provider notifies you (via webhook or API polling) → this is the DLR.

Common DLR statuses (simplified)

DELIVERED → Reached handset (good).

SENT → Left provider, en route (intermediate).

FAILED / UNDELIVERED → Could not be delivered (bad).

EXPIRED → Timed out in carrier network.

REJECTED → Carrier rejected (e.g., invalid number, blocked).

UNKNOWN → Carrier couldn’t give final status.

Finality: Some statuses are final (delivered, failed, undelivered), some are transient (queued, sent).

Retries: For transient network errors, you might retry. For hard errors (invalid number), don’t.

DLQ: Failed messages after max retries → Dead Letter Queue for later inspection

Inbound Webhooks

SMS isn’t just outbound (you → user). Users can also reply back to your SMS or send new ones to your number.
(Not necessary for now I guess)
