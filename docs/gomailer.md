Simple Mail transfer Protocol is a simple Protocol used to send email.(only to send)
Recieveing is done by POP or IMAP
How does SMTP work?
When you click on send mail it sends to smtp server for Example: smtp.gmail.com
then gmail sends it to reciever mail server. WHich stays in the server.
Until the logs in and downloads the mail using POP/IMAP
SMTP uses TCP protocol.

SSL/TLS(secure shell Layer/Transport layer Security):

Is a ENcrytion handshake for website security
a 3 way handshkae happens when the client coonects to server
by sending and establishing the security certificate
So why TLS,
If you want to send MAIL through smtp you need make it secure
Two main ways to add TLS
Port 465 → “implicit TLS”
As soon as you connect, you go straight into the tunnel (TLS).
Only after the lock is in place do you start the email conversation
ClientHello → ServerHello → Certificate → key exchange → Finished

Port 587 → “STARTTLS”
First, you say hello in plain text:
Server: Hi, I can do STARTTLS if you want.
Then you say: Okay, let’s switch to TLS.
Now the tunnel is built, and you continue securely.
Problem: a hacker sitting in the middle could delete the “I can do STARTTLS” message. If your client isn’t strict, it might continue without TLS

Why Gmail blocks raw passwords?

In the old days, apps (like Thunderbird, Outlook, or phone mail apps) just asked for your email + password and sent it to Gmail.

That’s risky because:

If the app is hacked, it can steal your Google password.
If the connection isn’t fully secure, your password could leak.
That password isn’t just for email — it unlocks all of Google (Drive, Photos, etc.).
To reduce that risk, Google blocks “less secure apps” from using your real Google password.

How app passwords fix this?

An app password is a special 16-character code you generate in your Google Account.
It works like a one-purpose key:
Only valid for logging into Gmail via old-style login (IMAP/SMTP/POP).
It does not work for Google Drive, YouTube, Calendar, etc.
You can revoke it anytime without changing your main password.
So if the app (or your device) gets hacked, the attacker only gets the app password, which you can delete — your main Google password stays safe.

Why deliverability matters

When you send an email, two things happen:
The recipient’s mail server (like Gmail or Outlook) has to accept it.
The recipient’s spam filter has to decide Inbox vs. Spam folder vs. Rejection.
Providers use lots of signals to decide this — mainly authentication, sender reputation, and content.

THe standard followed is SPF, DKIM, DMARC
SPF = “who can send,” DKIM = “prove it wasn’t altered(through crytographic key),” DMARC = “what to do if something fails.”

email is reputation-based networking:

If you’re a hobby sender, Gmail SMTP works — nobody cares if a test message lands in spam occasionally.

If you’re a production app, deliverability is part of your product. You need the right infrastructure
