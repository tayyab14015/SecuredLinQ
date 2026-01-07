import { useState, useEffect, useCallback } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { loadsApi, meetingsApi, smsApi, mediaApi } from '../../api/client';
import type { Load, MediaItem } from '../../api/client';

export default function LoadDetail() {
  const { guestRand } = useParams<{ guestRand: string }>();
  const navigate = useNavigate();
  
  const [load, setLoad] = useState<Load | null>(null);
  const [media, setMedia] = useState<MediaItem[]>([]);
  const [meetingRoom, setMeetingRoom] = useState<{ room_id: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [sendingSMS, setSendingSMS] = useState(false);
  const [smsStatus, setSmsStatus] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  const fetchData = useCallback(async () => {
    if (!guestRand) return;
    
    try {
      setLoading(true);
      const [loadData, mediaData] = await Promise.all([
        loadsApi.getByGuestRand(guestRand),
        mediaApi.getLoadMedia(guestRand).catch(() => []),
      ]);
      setLoad(loadData);
      setMedia(mediaData);
    } catch (error) {
      console.error('Failed to fetch load data:', error);
    } finally {
      setLoading(false);
    }
  }, [guestRand]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreateMeeting = async () => {
    if (!guestRand) return;
    
    try {
      const room = await meetingsApi.create(guestRand);
      setMeetingRoom({ room_id: room.room_id });
    } catch (error) {
      console.error('Failed to create meeting:', error);
    }
  };

  const handleSendLink = async () => {
    if (!load?.guest_phone || !meetingRoom?.room_id) return;
    
    try {
      setSendingSMS(true);
      setSmsStatus(null);
      
      const meetingUrl = `${window.location.origin}/join/${meetingRoom.room_id}`;
      await smsApi.sendMeetingLink(load.guest_phone, meetingUrl);
      
      setSmsStatus({ type: 'success', message: 'Meeting link sent successfully!' });
    } catch (error) {
      setSmsStatus({ type: 'error', message: 'Failed to send SMS. Please try again.' });
    } finally {
      setSendingSMS(false);
    }
  };

  const handleJoinCall = () => {
    if (meetingRoom?.room_id) {
      navigate(`/admin/meeting/${meetingRoom.room_id}`);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--bg-primary)]">
        <div className="spinner"></div>
      </div>
    );
  }

  if (!load) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-[var(--bg-primary)]">
        <p className="text-slate-600 mb-4">Load not found.</p>
        <Link to="/admin/dashboard" className="btn btn-primary">
          Back to Dashboard
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      {/* Header */}
      <header className="bg-white border-b border-[var(--border)]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-4">
              <Link
                to="/admin/dashboard"
                className="text-slate-500 hover:text-slate-700 transition-colors"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </Link>
              <div>
                <h1 className="text-lg font-semibold text-slate-800">Load Details</h1>
                <p className="text-xs text-slate-500">#{load.load_number}</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Load info */}
          <div className="lg:col-span-1 space-y-6">
            <div className="card">
              <h2 className="text-lg font-semibold text-slate-800 mb-4">Driver Information</h2>
              <dl className="space-y-3">
                <div>
                  <dt className="text-sm text-slate-500">Name</dt>
                  <dd className="font-medium">{load.guest_first_name} {load.guest_last_name}</dd>
                </div>
                <div>
                  <dt className="text-sm text-slate-500">Phone</dt>
                  <dd className="font-medium">{load.guest_phone || 'N/A'}</dd>
                </div>
                <div>
                  <dt className="text-sm text-slate-500">Email</dt>
                  <dd className="font-medium">{load.guest_email || 'N/A'}</dd>
                </div>
                <div>
                  <dt className="text-sm text-slate-500">Load Number</dt>
                  <dd className="font-mono bg-slate-100 px-2 py-1 rounded inline-block">
                    {load.load_number || 'N/A'}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm text-slate-500">Trailer Number</dt>
                  <dd className="font-medium">{load.trailer_number || 'N/A'}</dd>
                </div>
              </dl>
            </div>

            {/* Meeting controls */}
            <div className="card">
              <h2 className="text-lg font-semibold text-slate-800 mb-4">Video Call</h2>
              
              {!meetingRoom ? (
                <button onClick={handleCreateMeeting} className="btn btn-primary w-full">
                  Create Meeting Room
                </button>
              ) : (
                <div className="space-y-3">
                  <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
                    <p className="text-sm text-green-700 font-medium">Meeting room created!</p>
                    <p className="text-xs text-green-600 mt-1 font-mono break-all">
                      Room ID: {meetingRoom.room_id}
                    </p>
                  </div>
                  
                  <button onClick={handleJoinCall} className="btn btn-accent w-full">
                    Join Video Call
                  </button>
                  
                  <button 
                    onClick={handleSendLink}
                    disabled={sendingSMS || !load.guest_phone}
                    className="btn btn-outline w-full"
                  >
                    {sendingSMS ? 'Sending...' : 'Send Link via SMS'}
                  </button>
                  
                  {smsStatus && (
                    <div className={`p-3 rounded-lg text-sm ${
                      smsStatus.type === 'success' 
                        ? 'bg-green-50 text-green-700 border border-green-200' 
                        : 'bg-red-50 text-red-700 border border-red-200'
                    }`}>
                      {smsStatus.message}
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Media gallery */}
          <div className="lg:col-span-2">
            <div className="card">
              <h2 className="text-lg font-semibold text-slate-800 mb-4">Media Gallery</h2>
              
              {media.length === 0 ? (
                <div className="text-center py-12 text-slate-500">
                  <svg className="w-12 h-12 mx-auto mb-3 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  <p>No media files yet.</p>
                  <p className="text-sm mt-1">Screenshots and recordings will appear here.</p>
                </div>
              ) : (
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
                  {media.map((item, index) => (
                    <a
                      key={item.key || index}
                      href={item.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="group relative aspect-video bg-slate-100 rounded-lg overflow-hidden hover:ring-2 hover:ring-[var(--primary)] transition-all"
                    >
                      {item.type === 'screenshot' ? (
                        <img
                          src={item.url}
                          alt={`Screenshot ${index + 1}`}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center bg-slate-800">
                          <svg className="w-8 h-8 text-white" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M8 5v14l11-7z" />
                          </svg>
                        </div>
                      )}
                      <div className="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center">
                        <span className="opacity-0 group-hover:opacity-100 text-white text-sm font-medium transition-opacity">
                          View
                        </span>
                      </div>
                      <span className={`absolute top-2 right-2 px-2 py-0.5 rounded text-xs font-medium ${
                        item.type === 'screenshot' ? 'bg-blue-500 text-white' : 'bg-red-500 text-white'
                      }`}>
                        {item.type === 'screenshot' ? 'Screenshot' : 'Video'}
                      </span>
                    </a>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

