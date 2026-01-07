import { useEffect, useState, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { meetingsApi } from '../api/client';
import type { MeetingRoom as MeetingRoomType } from '../api/client';
import VideoCallInterface from '../components/VideoCallInterface';

interface MeetingRoomProps {
  isAdmin?: boolean;
}

export default function MeetingRoom({ isAdmin = false }: MeetingRoomProps) {
  const { roomId } = useParams<{ roomId: string }>();
  
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [meetingRoom, setMeetingRoom] = useState<MeetingRoomType | null>(null);
  const [callEnded, setCallEnded] = useState(false);

  useEffect(() => {
    const fetchMeetingRoom = async () => {
      if (!roomId || roomId === 'undefined' || roomId === 'null') {
        setError('No room ID provided.');
        setLoading(false);
        return;
      }

      try {
        const room = await meetingsApi.getByRoomId(roomId);
        
        if (room.status === 'ended') {
          setCallEnded(true);
        }
        
        setMeetingRoom(room);
      } catch (err) {
        console.error('Error fetching meeting room:', err);
        setError('Unable to find the meeting room. It may have been closed.');
      } finally {
        setLoading(false);
      }
    };

    fetchMeetingRoom();
  }, [roomId]);

  const handleCallEnded = useCallback(() => {
    setCallEnded(true);
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-slate-900 p-4">
        <div className="spinner mb-4"></div>
        <p className="text-white text-lg">Connecting to meeting...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-slate-900 p-4">
        <div className="bg-white/95 backdrop-blur-sm rounded-2xl shadow-2xl p-8 max-w-md w-full text-center">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h1 className="text-xl font-bold text-slate-800 mb-2">Meeting Not Found</h1>
          <p className="text-slate-600 mb-6">{error}</p>
          {isAdmin ? (
            <Link to="/admin/dashboard" className="btn btn-primary">
              Back to Dashboard
            </Link>
          ) : (
            <p className="text-sm text-slate-500">Please contact support if you need assistance.</p>
          )}
        </div>
      </div>
    );
  }

  if (callEnded) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-slate-900 p-4">
        <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full mx-4 animate-fade-in overflow-hidden">
          <div className="p-8 text-center">
            <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-6">
              <svg className="w-10 h-10 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h1 className="text-2xl font-bold text-slate-800 mb-3">Call Ended</h1>
            <p className="text-slate-600 mb-8">
              {isAdmin 
                ? 'The video call has ended. All recordings have been saved.'
                : 'Thank you for completing the video verification.'}
            </p>
            {isAdmin ? (
              <Link 
                to="/admin/dashboard" 
                className="inline-block w-full bg-slate-800 hover:bg-slate-900 text-white font-medium py-3 px-6 rounded-lg transition-colors text-center"
              >
                Back to Dashboard
              </Link>
            ) : (
              <p className="text-sm text-slate-500">
                You may now close this window.
              </p>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-900">
      {/* Header */}
      <div className="absolute top-0 left-0 right-0 z-10 bg-gradient-to-b from-black/50 to-transparent p-4">
        <div className="flex items-center justify-between max-w-7xl mx-auto">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-white/10 backdrop-blur-sm rounded-xl flex items-center justify-center">
              <svg 
                className="w-5 h-5 text-amber-400" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path 
                  strokeLinecap="round" 
                  strokeLinejoin="round" 
                  strokeWidth={2} 
                  d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" 
                />
              </svg>
            </div>
            <div>
              <h1 className="text-white font-semibold">SecuredLinQ</h1>
              <p className="text-white/60 text-xs">
                {isAdmin ? 'Admin View' : 'Video Call'}
              </p>
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            <span className="flex items-center gap-1.5 px-3 py-1.5 bg-green-500/20 border border-green-500/30 rounded-full text-green-400 text-sm">
              <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
              Live
            </span>
          </div>
        </div>
      </div>

      {/* Video call interface */}
      {meetingRoom && (
        <VideoCallInterface
          roomId={meetingRoom.roomId || meetingRoom.room_id || ''}
          channelName={meetingRoom.channelName || meetingRoom.channel_name || ''}
          guestRand={meetingRoom.guest_rand || meetingRoom.roomId || ''}
          loadId={meetingRoom.load_id || meetingRoom.loadId || 0}
          isAdmin={isAdmin}
          onCallEnded={handleCallEnded}
        />
      )}
    </div>
  );
}

