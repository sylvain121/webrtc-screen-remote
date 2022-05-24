package WebRTC

import (
	"context"
	"fmt"
	"github.com/pion/webrtc/v3"
	"os"
)

type WebRtcEndpoint struct {
	VideoTrack *webrtc.TrackLocalStaticSample
	cursor     *webrtc.DataChannel
}

func (endpoint *WebRtcEndpoint) NewConnection(offer webrtc.SessionDescription) (webrtc.SessionDescription, context.Context) {

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})

	if err != nil {
		fmt.Println("Error while NewPeerConnection")
		panic(err)
	}
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if videoTrackErr != nil {
		fmt.Println("Error while NewTackLocalStaticSample")
		panic(videoTrackErr)
	}

	endpoint.VideoTrack = videoTrack
	rtpSender, videoTrackErr := peerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		fmt.Println("Error while AddTrack")
		panic(videoTrackErr)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}
	})

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		fmt.Println("Error while SetRemoteDescription")
		panic(err)
	}
	answer, err := peerConnection.CreateAnswer(nil)

	if err != nil {
		fmt.Println("Error while CreateAnswer")
		panic(err)
	}

	if err := peerConnection.SetLocalDescription(answer); err != nil {
		fmt.Println("Error while SetLocalDescription")
		panic(err)
	}

	endpoint.cursor, err = peerConnection.CreateDataChannel("cursor", nil)
	if err != nil {
		panic(err)
	}

	return *peerConnection.LocalDescription(), iceConnectedCtx
}
