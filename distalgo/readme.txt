Tested killing a Head , Tail and some intermittent server

Killing head will send an update request to client to update the new head

Killing an intermittent server will send an update request to previous server to point to next server of the killed server

Killing a tail sends a request to previous server to make it a tail and client to update the chain tail

every time out os read from configuration
