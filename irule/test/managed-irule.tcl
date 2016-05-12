when HTTP_REQUEST {
  switch -glob [string tolower [HTTP::path]] {
    "/status" {
    	HTTP::respond 200 content "200 OK"
    }
    default {
    	# Do nothing
    }
  }
}