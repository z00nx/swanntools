package main

import (
	"crypto/tls"
	"net"
	"bufio"
	"strconv"
	"bytes"
	log "github.com/Sirupsen/logrus"
)

const (
	SuccessfulAuthString = "200"
	FailedAuthString     = "403"
	InvalidChannelString = "400"
	ChannelInUseString   = "409"
)

func StartListener() {
	// Load server key pair
	cert, err := tls.LoadX509KeyPair(config.certs+"/server.pem", config.certs+"/server.key")
	if err != nil {
		log.Fatalln("Unable to load server key pair: ", err.Error())
	}

	// Add certificate to TLS config
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// Listen on the bindAddr for stream bytes
	listener, err := tls.Listen("tcp", config.bindAddr.String(), tlsConfig)
	if err != nil {
		log.Fatalln("Unable to start TLS listener: ", err.Error())
	}

	log.Infof("Server ready and listening on: %s", config.bindAddr)

	for {
		// TODO: Use Mutexes to protect channels from simultaneous writes

		// Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			log.Warnln("An error occured when accepting a connection: ", err.Error())
			// Close the connection on error
			conn.Close()
			// Listen for more connections
			continue
		}

		// Handle the connection
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	var isAuthenticated bool = false // isAuthenticated stores whether the client is authenticated
	var channel int                  // channel stores the channel the client is sending
	var response string              // response stores the response code to send to the client

	defer func(channel *int) {
		// Close the connection upon connection end
		conn.Close()
		// Remove channel from channelsInUse if appropriate
		if channel != nil {
			if pos, isPresent := intPositionInSlice(channel, &channelsInUse); isPresent {
				channelsInUse = append(channelsInUse[:pos], channelsInUse[pos+1:]...)
			}
		}
	}(&channel)

	// Read the authentication data provided by the client
	authData := bufio.NewReader(conn)

	// Attempt authentication
	isAuthenticated, channel, response = parseAuthMessage(authData)

	log.WithFields(log.Fields{"source": conn.RemoteAddr().String(), "channel": channel, "code": response, }).
		Infof("Auth status: %v\n", isAuthenticated)

	// Send the response to the client
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Warnln("Unable to write response to client: ", err.Error())
	}

	// Cease execution if authentication failed
	if !isAuthenticated {
		return
	}

	// Append the channel to slice of channels in use
	if intInSlice(&channel, &channelsInUse) {
		channelsInUse = append(channelsInUse, channel)
	}

	// Get the camera stream
	for {
		// Create a byte array to store data
		data := make([]byte, socketBufferSize)

		// Read the data from the connection
		n, err := conn.Read(data)
		if err != nil {
			log.WithFields(log.Fields{
				"source": conn.RemoteAddr().String(), "channel": channel,
			}).Warnf("An error occurred while reading stream: %s", err.Error())
			break
		}

		// Send data to each consumer
		for _, consumer := range config.consumers {
			consumer.Receiver <- Data{channel, data[:n]}
		}
	}
}

func parseAuthMessage(r *bufio.Reader) (isAuthenticated bool, channelNum int, responseCode string) {
	var nilInt int

	// Parse channel and password
	msg, err := r.ReadString('\n')
	if err != nil {
		log.Warnln("Unable to retrieve authentication message: ", err.Error())
		return false, nilInt, FailedAuthString
	}
	channelInput := string(msg[0])
	// Ensure that line break is removed
	passwordInput := string(bytes.Trim([]byte(msg[1:]), "\x0a"))

	// Validate length, accounting for the line break
	if len(msg) < 3 {
		log.Warnln("Authentication failed due to invalid authentication message length")
		return false, nilInt, FailedAuthString
	}

	// Validate channel
	intChannel, err := strconv.Atoi(channelInput)
	if len(channelsInUse) >= maxChannels {
		log.Warnln("You cannot have greater than %d streams", maxChannels)
		return false, nilInt, InvalidChannelString
	} else if err != nil || intChannel > maxChannels {
		log.Warnln("All channels need to be a number between 1 and %d", maxChannels)
		return false, nilInt, InvalidChannelString
	} else if intInSlice(&intChannel, &channelsInUse) {
		log.Warnln("The channel %d is currently receiving a stream", intChannel)
		return false, nilInt, ChannelInUseString
	}

	// Validate password
	if passwordInput != config.key {
		return false, nilInt, FailedAuthString
	}

	return true, intChannel, SuccessfulAuthString
}
