import { useEffect, useRef, useState, useCallback } from 'react';
import AgoraRTC from 'agora-rtc-sdk-ng';
import type {
  IAgoraRTCClient,
  IAgoraRTCRemoteUser,
  ICameraVideoTrack,
  IMicrophoneAudioTrack,
} from 'agora-rtc-sdk-ng';
import { agoraApi, mediaApi, meetingsApi } from '../api/client';

// Agora App ID from environment
const AGORA_APP_ID = import.meta.env.VITE_AGORA_APP_ID || '';

interface VideoCallInterfaceProps {
  roomId: string;
  channelName: string;
  guestRand?: string;
  loadId?: number;
  isAdmin?: boolean;
  onCallEnded?: () => void;
}

export default function VideoCallInterface({
  roomId,
  channelName,
  guestRand,
  loadId,
  isAdmin = false,
  onCallEnded,
}: VideoCallInterfaceProps) {
  // Refs
  const clientRef = useRef<IAgoraRTCClient | null>(null);
  const localVideoRef = useRef<HTMLDivElement>(null);
  const remoteVideoRef = useRef<HTMLDivElement>(null);
  const localTracksRef = useRef<{
    audioTrack: IMicrophoneAudioTrack | null;
    videoTrack: ICameraVideoTrack | null;
  }>({ audioTrack: null, videoTrack: null });
  const isJoiningRef = useRef(false); // Track if we're currently joining

  // State
  const [joined, setJoined] = useState(false);
  const [remoteUser, setRemoteUser] = useState<IAgoraRTCRemoteUser | null>(null);
  const [isAudioMuted, setIsAudioMuted] = useState(false);
  const [isVideoMuted, setIsVideoMuted] = useState(false);
  const [isRecording, setIsRecording] = useState(false);
  const [recordingData, setRecordingData] = useState<{ resourceId: string; sid: string; uid: string; cname: string } | null>(null);
  const [cameras, setCameras] = useState<MediaDeviceInfo[]>([]);
  const [currentCameraIndex, setCurrentCameraIndex] = useState(0);
  const [connectionState, setConnectionState] = useState<'DISCONNECTED' | 'CONNECTING' | 'CONNECTED' | 'RECONNECTING'>('DISCONNECTED');

  // Generate unique UID as number (will be converted to string for Agora)
  const uid = useRef(Math.floor(Math.random() * 100000));

  // Initialize Agora client
  useEffect(() => {
    const client = AgoraRTC.createClient({ mode: 'rtc', codec: 'vp8' });
    clientRef.current = client;

    // Event handlers
    client.on('user-published', async (user, mediaType) => {
      await client.subscribe(user, mediaType);
      
      if (mediaType === 'video') {
        setRemoteUser(user);
        if (remoteVideoRef.current) {
          user.videoTrack?.play(remoteVideoRef.current, { fit: 'contain' });
        }
      }
      
      if (mediaType === 'audio') {
        user.audioTrack?.play();
      }
    });

    client.on('user-unpublished', (_user, mediaType) => {
      if (mediaType === 'video') {
        setRemoteUser(null);
      }
    });

    client.on('user-left', () => {
      setRemoteUser(null);
    });

    client.on('connection-state-change', (curState) => {
      setConnectionState(curState as typeof connectionState);
    });

    // Get available cameras
    AgoraRTC.getCameras().then((deviceList) => {
      setCameras(deviceList);
    });

    return () => {
      leaveChannel();
    };
  }, []);

  // Join channel
  const joinChannel = useCallback(async () => {
    // Prevent multiple simultaneous join attempts
    if (!clientRef.current || joined || isJoiningRef.current) {
      return;
    }

    // Check if client is already in a connecting/connected state
    const connectionState = clientRef.current.connectionState;
    if (connectionState === 'CONNECTING' || connectionState === 'CONNECTED' || connectionState === 'RECONNECTING') {
      console.log('Client already in state:', connectionState);
      return;
    }

    isJoiningRef.current = true;

    try {
      setConnectionState('CONNECTING');

      // Get token from backend (includes appId)
      const tokenResponse = await agoraApi.getToken({
        channelName,
        uid: uid.current,
        role: 'publisher',
      });

      // Use appId from backend response, fallback to env variable
      const appId = tokenResponse.appId || AGORA_APP_ID || '';
      
      if (!appId) {
        throw new Error('Agora App ID is not configured. Please set VITE_AGORA_APP_ID environment variable.');
      }

      // Convert UID to string for Agora (required)
      const uidString = String(uid.current);

      // Double-check client state before joining
      if (clientRef.current.connectionState === 'CONNECTING' || 
          clientRef.current.connectionState === 'CONNECTED') {
        console.log('Client state changed during token fetch, aborting join');
        isJoiningRef.current = false;
        return;
      }
       
      await clientRef.current.join(appId, channelName, tokenResponse.token, uidString);

      // Create and publish local tracks
      const [audioTrack, videoTrack] = await AgoraRTC.createMicrophoneAndCameraTracks();
      localTracksRef.current = { audioTrack, videoTrack };

      // Play local video with contain fit
      if (localVideoRef.current) {
        videoTrack.play(localVideoRef.current, { fit: 'contain' });
      }

      // Publish tracks
      await clientRef.current.publish([audioTrack, videoTrack]);

      setJoined(true);
      setConnectionState('CONNECTED');
      isJoiningRef.current = false;


    } catch (error) {
      console.error('Failed to join channel:', error);
      setConnectionState('DISCONNECTED');
      isJoiningRef.current = false;
    }
  }, [channelName, roomId, guestRand, isAdmin]); // Removed 'joined' from dependencies

  // Auto-join on mount (only once)
  useEffect(() => {
    if (!joined && !isJoiningRef.current && clientRef.current) {
      joinChannel();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run once on mount

  // Leave channel
  const leaveChannel = async () => {
    // Reset joining flag
    isJoiningRef.current = false;
    
    if (localTracksRef.current.audioTrack) {
      localTracksRef.current.audioTrack.close();
    }
    if (localTracksRef.current.videoTrack) {
      localTracksRef.current.videoTrack.close();
    }
    localTracksRef.current = { audioTrack: null, videoTrack: null };

    if (clientRef.current) {
      await clientRef.current.leave();
    }

    setJoined(false);
    setRemoteUser(null);
  };

  // Toggle audio
  const toggleAudio = async () => {
    if (localTracksRef.current.audioTrack) {
      // If currently muted (isAudioMuted=true), enable the track
      // If not muted (isAudioMuted=false), disable the track
      const newMutedState = !isAudioMuted;
      await localTracksRef.current.audioTrack.setEnabled(!newMutedState);
      setIsAudioMuted(newMutedState);
    }
  };

  // Toggle video
  const toggleVideo = async () => {
    if (localTracksRef.current.videoTrack) {
      // If currently muted (isVideoMuted=true), enable the track
      // If not muted (isVideoMuted=false), disable the track
      const newMutedState = !isVideoMuted;
      await localTracksRef.current.videoTrack.setEnabled(!newMutedState);
      setIsVideoMuted(newMutedState);
      
      // Re-play the video if enabling
      if (!newMutedState && localVideoRef.current) {
        localTracksRef.current.videoTrack.play(localVideoRef.current, { fit: 'contain' });
      }
    }
  };

  // Switch camera (for drivers)
  const switchCamera = async () => {
    if (cameras.length <= 1 || !localTracksRef.current.videoTrack) return;

    const nextIndex = (currentCameraIndex + 1) % cameras.length;
    const nextCamera = cameras[nextIndex];

    try {
      await localTracksRef.current.videoTrack.setDevice(nextCamera.deviceId);
      setCurrentCameraIndex(nextIndex);
    } catch (error) {
      console.error('Failed to switch camera:', error);
    }
  };

  // Capture screenshot (admin only)
  const captureScreenshot = async () => {
    if (!remoteUser?.videoTrack || !guestRand) return;

    try {
      // Get video element
      const videoElement = remoteVideoRef.current?.querySelector('video');
      if (!videoElement) return;

      // Create canvas and capture frame
      const canvas = document.createElement('canvas');
      canvas.width = videoElement.videoWidth || 640;
      canvas.height = videoElement.videoHeight || 480;
      
      const ctx = canvas.getContext('2d');
      if (!ctx) return;
      
      ctx.drawImage(videoElement, 0, 0);
      const imageData = canvas.toDataURL('image/jpeg', 0.9);

      // Save to S3
      await mediaApi.saveScreenshot({
        imageData,
        roomId,
        loadId,
      });

      // Show success feedback (could add a toast here)
      console.log('Screenshot captured successfully');
    } catch (error) {
      console.error('Failed to capture screenshot:', error);
    }
  };

  // Start recording (admin only)
  const startRecording = async () => {
    if (isRecording || !roomId) return;

    try {
      // Get recording token
      const recordingUIDNum = uid.current + 1; // Different UID for recording
      const recordingUID = String(recordingUIDNum);
      const { token } = await agoraApi.getToken({
        channelName,
        uid: recordingUIDNum, // getToken expects number
        role: 'subscriber',
      });

      const result = await agoraApi.startRecording({
        roomId,
        channelName,
        uid: recordingUID, // startRecording expects string
        token,
      });

      // Save recording data to localStorage
      if (result.uid && result.sid && result.resourceId && result.cname) {
        localStorage.setItem('recordingUID', result.uid);
        localStorage.setItem('recordingSID', result.sid);
        localStorage.setItem('recordingResourceId', result.resourceId);
        localStorage.setItem('recordingChannelName', result.cname);
      }

      setRecordingData({
        resourceId: result.resourceId,
        sid: result.sid,
        uid: result.uid,
        cname: result.cname,
      });
      setIsRecording(true);
    } catch (error) {
      console.error('Failed to start recording:', error);
    }
  };

  // Stop recording (admin only)
  const stopRecording = async () => {
    if (!isRecording) return;

    try {
      // Try to get data from state first, then fallback to localStorage
      let resourceId = recordingData?.resourceId;
      let sid = recordingData?.sid;
      let channelNameToUse = recordingData?.cname || channelName;
      let uidToUse = recordingData?.uid;

      // If data not in state, try localStorage (for page refresh recovery)
      if (!resourceId || !sid || !uidToUse) {
        resourceId = localStorage.getItem('recordingResourceId') || '';
        sid = localStorage.getItem('recordingSID') || '';
        channelNameToUse = localStorage.getItem('recordingChannelName') || channelName;
        uidToUse = localStorage.getItem('recordingUID') || '';
      }

      // Validate all required fields exist
      if (!resourceId || !sid || !channelNameToUse || !uidToUse) {
        console.error('Missing recording data. Cannot stop recording.');
        alert('Recording data not found. Please start a new recording.');
        setIsRecording(false);
        setRecordingData(null);
        return;
      }

      await agoraApi.stopRecording({
        channelName: channelNameToUse,
        resourceId,
        sid,
        uid: uidToUse, // Must match the UID used to start recording
      });

      // Clear localStorage after successful stop
      localStorage.removeItem('recordingUID');
      localStorage.removeItem('recordingSID');
      localStorage.removeItem('recordingResourceId');
      localStorage.removeItem('recordingChannelName');

      setIsRecording(false);
      setRecordingData(null);
    } catch (error) {
      console.error('Failed to stop recording:', error);
      // Don't clear localStorage on error so user can retry
    }
  };

  // End call
  const handleEndCall = async () => {
    // Stop recording if active
    if (isRecording) {
      await stopRecording();
    }

    // Leave channel
    await leaveChannel();

    // End meeting (admin only)
    if (isAdmin) {
      await meetingsApi.end(roomId);
    }

    onCallEnded?.();
  };

  return (
    <div className="h-screen flex flex-col">
      {/* Video grid - in-frame layout */}
      <div className="flex-1 p-4 pt-20 overflow-hidden">
        <div className="h-full max-h-[calc(100vh-200px)] grid grid-cols-2 gap-4">
          {/* Remote video (other participant) */}
          <div className="relative bg-slate-800 rounded-2xl overflow-hidden shadow-xl border border-white/10">
            <div 
              ref={remoteVideoRef} 
              className="absolute inset-0 w-full h-full"
              style={{ 
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center'
              }}
            >
              {!remoteUser && (
                <div className="w-full h-full flex items-center justify-center text-white/60 bg-slate-800">
                  <div className="text-center">
                    <svg className="w-20 h-20 mx-auto mb-4 opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                    <p className="text-lg">Waiting for {isAdmin ? 'driver' : 'admin'} to join...</p>
                  </div>
                </div>
              )}
            </div>
            <div className="absolute bottom-4 left-4 px-3 py-1.5 bg-black/60 backdrop-blur-sm rounded-lg text-white text-sm font-medium z-10">
              {isAdmin ? 'Driver' : 'Admin'}
            </div>
          </div>

          {/* Local video (self) */}
          <div className="relative bg-slate-700 rounded-2xl overflow-hidden shadow-xl border border-white/10">
            <div 
              ref={localVideoRef} 
              className="absolute inset-0 w-full h-full"
              style={{ 
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center'
              }}
            >
              {isVideoMuted && (
                <div className="absolute inset-0 bg-slate-700 flex items-center justify-center z-10">
                  <div className="text-center text-white/60">
                    <svg className="w-20 h-20 mx-auto mb-3 opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      <line x1="3" y1="21" x2="21" y2="3" strokeWidth={2} />
                    </svg>
                    <p className="text-lg">Camera Off</p>
                  </div>
                </div>
              )}
            </div>
            <div className="absolute bottom-4 left-4 px-3 py-1.5 bg-black/60 backdrop-blur-sm rounded-lg text-white text-sm font-medium z-10">
              You
            </div>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="bg-slate-800/90 backdrop-blur-sm border-t border-white/10 p-4">
        <div className="flex items-center justify-center gap-3 w-full">
          {/* Mute/unmute audio */}
          <button
            onClick={toggleAudio}
            className={`w-12 h-12 rounded-full flex items-center justify-center transition-colors ${
              isAudioMuted 
                ? 'bg-red-500 hover:bg-red-600' 
                : 'bg-white/10 hover:bg-white/20'
            }`}
            title={isAudioMuted ? 'Unmute' : 'Mute'}
          >
            {isAudioMuted ? (
              <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" />
              </svg>
            ) : (
              <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
              </svg>
            )}
          </button>

          {/* Toggle video */}
          <button
            onClick={toggleVideo}
            className={`w-12 h-12 rounded-full flex items-center justify-center transition-colors ${
              isVideoMuted 
                ? 'bg-red-500 hover:bg-red-600' 
                : 'bg-white/10 hover:bg-white/20'
            }`}
            title={isVideoMuted ? 'Turn on camera' : 'Turn off camera'}
          >
            {isVideoMuted ? (
              <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                <line x1="3" y1="21" x2="21" y2="3" stroke="currentColor" strokeWidth={2} />
              </svg>
            ) : (
              <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
            )}
          </button>

          {/* Switch camera (driver only) */}
          {!isAdmin && cameras.length > 1 && (
            <button
              onClick={switchCamera}
              className="w-12 h-12 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center transition-colors"
              title="Switch camera"
            >
              <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
          )}

          {/* Admin controls */}
          {isAdmin && (
            <>
              {/* Screenshot */}
              <button
                onClick={captureScreenshot}
                disabled={!remoteUser}
                className="w-12 h-12 rounded-full bg-blue-500 hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center transition-colors"
                title="Capture screenshot"
              >
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </button>

              {/* Recording */}
              <button
                onClick={isRecording ? stopRecording : startRecording}
                className={`w-12 h-12 rounded-full flex items-center justify-center transition-colors ${
                  isRecording 
                    ? 'bg-red-500 hover:bg-red-600 animate-pulse' 
                    : 'bg-amber-500 hover:bg-amber-600'
                }`}
                title={isRecording ? 'Stop recording' : 'Start recording'}
              >
                {isRecording ? (
                  <svg className="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
                    <rect x="6" y="6" width="12" height="12" rx="1" />
                  </svg>
                ) : (
                  <svg className="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
                    <circle cx="12" cy="12" r="6" />
                  </svg>
                )}
              </button>
            </>
          )}

          {/* End call */}
          <button
            onClick={handleEndCall}
            className="w-14 h-14 rounded-full bg-red-500 hover:bg-red-600 flex items-center justify-center transition-colors ml-4"
            title="End call"
          >
            <svg className="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 8l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2M5 3a2 2 0 00-2 2v1c0 8.284 6.716 15 15 15h1a2 2 0 002-2v-3.28a1 1 0 00-.684-.948l-4.493-1.498a1 1 0 00-1.21.502l-1.13 2.257a11.042 11.042 0 01-5.516-5.517l2.257-1.128a1 1 0 00.502-1.21L9.228 3.683A1 1 0 008.279 3H5z" />
            </svg>
          </button>
        </div>

        {/* Recording indicator */}
        {isRecording && (
          <div className="mt-3 flex items-center justify-center gap-2 text-red-400 text-sm">
            <span className="w-2 h-2 bg-red-400 rounded-full animate-pulse"></span>
            Recording in progress
          </div>
        )}

        {/* Connection status */}
        {connectionState !== 'CONNECTED' && (
          <div className="mt-3 flex items-center justify-center gap-2 text-amber-400 text-sm">
            <div className="w-4 h-4 border-2 border-amber-400/30 border-t-amber-400 rounded-full animate-spin"></div>
            {connectionState === 'CONNECTING' ? 'Connecting...' : 
             connectionState === 'RECONNECTING' ? 'Reconnecting...' : 'Disconnected'}
          </div>
        )}
      </div>
    </div>
  );
}

