import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { adminLoadsApi, adminDriversApi, meetingsApi, emailApi } from '../../api/client';
import type { Load, Driver } from '../../api/client';

type TabType = 'loads' | 'drivers';

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState<TabType>('loads');
  const [loads, setLoads] = useState<Load[]>([]);
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showAssignModal, setShowAssignModal] = useState<Load | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [showLinkModal, setShowLinkModal] = useState<{ load: Load; link: string; roomId: string } | null>(null);
  const [sendingEmail, setSendingEmail] = useState(false);
  const [emailSent, setEmailSent] = useState(false);
  
  const { logout, username } = useAuth();
  const navigate = useNavigate();

  const fetchLoads = useCallback(async () => {
    try {
      const data = await adminLoadsApi.getAll(1, 100);
      setLoads(data.loads || []);
    } catch (error) {
      console.error('Failed to fetch loads:', error);
    }
  }, []);

  const fetchDrivers = useCallback(async () => {
    try {
      const data = await adminDriversApi.getAll(1, 100);
      setDrivers(data.drivers || []);
    } catch (error) {
      console.error('Failed to fetch drivers:', error);
    }
  }, []);

  const fetchData = useCallback(async () => {
    setLoading(true);
    await Promise.all([fetchLoads(), fetchDrivers()]);
    setLoading(false);
  }, [fetchLoads, fetchDrivers]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Auto-refresh for real-time status updates
  useEffect(() => {
    const interval = setInterval(fetchLoads, 5000); // Refresh every 5 seconds
    return () => clearInterval(interval);
  }, [fetchLoads]);

  const handleLogout = async () => {
    await logout();
    navigate('/admin/login', { replace: true });
  };

  const handleStartMeeting = async (load: Load) => {
    try {
      // Mark meeting as started
      await adminLoadsApi.startMeeting(load.id);
      // Create meeting room
      const room = await meetingsApi.create(undefined, load.id);
      const roomId = room.roomId || room.room_id || '';
      
      // Get the meeting link for driver
      const driverLink = `${window.location.origin}/join/${roomId}`;
      
      // Reset email sent state
      setEmailSent(false);
      
      // Show modal with copyable link
      setShowLinkModal({ load, link: driverLink, roomId });
    } catch (error) {
      console.error('Failed to start meeting:', error);
      alert('Failed to start meeting. Please try again.');
    }
  };
  
  const handleSendEmail = async () => {
    if (!showLinkModal?.load.driver) {
      alert('No driver assigned to this load');
      return;
    }
    
    const driver = showLinkModal.load.driver;
    const driverEmail = driver.email;
    
    if (!driverEmail) {
      alert('Driver does not have an email address. Please update their profile.');
      return;
    }
    
    setSendingEmail(true);
    try {
      await emailApi.sendMeetingLink(
        driverEmail,
        `${driver.first_name} ${driver.last_name}`,
        showLinkModal.link,
        showLinkModal.load.load_number
      );
      setEmailSent(true);
      alert('Meeting link sent to driver\'s email successfully!');
    } catch (error) {
      console.error('Failed to send email:', error);
      alert('Failed to send email. Please try again or copy the link manually.');
    } finally {
      setSendingEmail(false);
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      alert('Link copied to clipboard!');
    } catch (err) {
      console.error('Failed to copy:', err);
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.opacity = '0';
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      alert('Link copied to clipboard!');
    }
  };

  const proceedToMeeting = () => {
    if (showLinkModal) {
      setShowLinkModal(null);
      navigate(`/admin/meeting/${showLinkModal.roomId}`);
    }
  };

  const filteredLoads = statusFilter === 'all' 
    ? loads 
    : loads.filter(l => l.status === statusFilter);

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      'Unassigned': 'bg-slate-100 text-slate-700',
      'Assigned': 'bg-blue-100 text-blue-700',
      'Completed': 'bg-green-100 text-green-700',
    };
    return styles[status] || 'bg-slate-100 text-slate-700';
  };

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      {/* Header */}
      <header className="bg-white border-b border-[var(--border)] sticky top-0 z-10">
        <div className="w-full px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-slate-700 to-slate-900 rounded-xl flex items-center justify-center">
                <svg 
                  className="w-5 h-5 text-amber-500" 
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
                <h1 className="text-lg font-semibold text-slate-800">SecuredLinQ</h1>
                <p className="text-xs text-slate-500">Admin Dashboard</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <span className="text-sm text-slate-600">
                Welcome, <span className="font-medium">{username}</span>
              </span>
              <button
                onClick={handleLogout}
                className="btn btn-outline text-sm"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="w-full px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats */}
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4 mb-8">
          <div className="card">
            <p className="text-sm text-slate-500 mb-1">Total Loads</p>
            <p className="text-2xl font-bold text-slate-800">{loads.length}</p>
          </div>
          <div className="card">
            <p className="text-sm text-slate-500 mb-1">Unassigned</p>
            <p className="text-2xl font-bold text-slate-600">
              {loads.filter(l => l.status === 'Unassigned').length}
            </p>
          </div>
          <div className="card">
            <p className="text-sm text-slate-500 mb-1">Assigned</p>
            <p className="text-2xl font-bold text-blue-600">
              {loads.filter(l => l.status === 'Assigned').length}
            </p>
          </div>
          <div className="card">
            <p className="text-sm text-slate-500 mb-1">Completed</p>
            <p className="text-2xl font-bold text-green-600">
              {loads.filter(l => l.status === 'Completed').length}
            </p>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-6">
          <button
            onClick={() => setActiveTab('loads')}
            className={`px-4 py-2 rounded-lg font-medium text-sm transition-colors ${
              activeTab === 'loads'
                ? 'bg-slate-800 text-white'
                : 'bg-white text-slate-600 hover:bg-slate-50'
            }`}
          >
            Load Management
          </button>
          <button
            onClick={() => setActiveTab('drivers')}
            className={`px-4 py-2 rounded-lg font-medium text-sm transition-colors ${
              activeTab === 'drivers'
                ? 'bg-slate-800 text-white'
                : 'bg-white text-slate-600 hover:bg-slate-50'
            }`}
          >
            Registered Drivers ({drivers.length})
          </button>
        </div>

        {activeTab === 'loads' && (
          <>
            {/* Actions bar */}
            <div className="flex flex-col sm:flex-row gap-4 mb-6">
              <div className="flex gap-2">
                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="input w-40"
                >
                  <option value="all">All Status</option>
                  <option value="Unassigned">Unassigned</option>
                  <option value="Assigned">Assigned</option>
                  <option value="Completed">Completed</option>
                </select>
              </div>
              <div className="flex gap-2 sm:ml-auto">
                <button
                  onClick={fetchLoads}
                  className="btn btn-outline flex items-center gap-2"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Refresh
                </button>
                <button
                  onClick={() => setShowCreateModal(true)}
                  className="btn btn-primary flex items-center gap-2"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Create Load
                </button>
              </div>
            </div>

            {/* Loads table */}
            <div className="card p-0">
              {loading ? (
                <div className="flex items-center justify-center py-12">
                  <div className="spinner"></div>
                </div>
              ) : filteredLoads.length === 0 ? (
                <div className="text-center py-12 text-slate-500">
                  No loads found.
                </div>
              ) : (
                <div className="table-container">
                  <table className="table">
                    <thead>
                      <tr>
                        <th>Load #</th>
                        <th>Status</th>
                        <th>Driver</th>
                        <th>Description</th>
                        <th>Created</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredLoads.map((load) => (
                        <tr key={load.id}>
                          <td>
                            <span className="font-mono font-medium">{load.load_number}</span>
                          </td>
                          <td>
                            <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusBadge(load.status)}`}>
                              {load.status}
                            </span>
                          </td>
                          <td>
                            {load.driver ? (
                              <div>
                                <p className="font-medium text-slate-800">
                                  {load.driver.first_name} {load.driver.last_name}
                                </p>
                                <p className="text-xs text-slate-500">{load.driver.phone_number}</p>
                              </div>
                            ) : (
                              <span className="text-slate-400">Not assigned</span>
                            )}
                          </td>
                          <td className="max-w-xs truncate">
                            {load.description || '-'}
                          </td>
                          <td className="text-sm text-slate-500">
                            {new Date(load.created_at).toLocaleDateString()}
                          </td>
                          <td>
                            <div className="flex gap-2 flex-wrap">
                              {load.status === 'Unassigned' && (
                                <button
                                  onClick={() => setShowAssignModal(load)}
                                  className="text-blue-600 hover:text-blue-800 font-medium text-sm"
                                >
                                  Assign
                                </button>
                              )}
                              {load.status === 'Completed' && (
                                <button
                                  onClick={() => handleStartMeeting(load)}
                                  className="btn bg-emerald-600 hover:bg-emerald-700 text-white text-xs px-3 py-1.5 flex items-center gap-1"
                                >
                                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                  </svg>
                                  Start Meeting
                                </button>
                              )}
                              <button
                                onClick={() => navigate(`/admin/gallery/${load.id}`)}
                                className="btn bg-purple-600 hover:bg-purple-700 text-white text-xs px-3 py-1.5 flex items-center gap-1"
                                title="View Screenshots"
                              >
                                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                </svg>
                                Gallery
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </>
        )}

        {activeTab === 'drivers' && (
          <div className="card p-0">
            {loading ? (
              <div className="flex items-center justify-center py-12">
                <div className="spinner"></div>
              </div>
            ) : drivers.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                No registered drivers yet.
              </div>
            ) : (
              <div className="table-container">
                <table className="table">
                  <thead>
                    <tr>
                      <th>Driver</th>
                      <th>Username</th>
                      <th>Phone</th>
                      <th>Status</th>
                      <th>Registered</th>
                    </tr>
                  </thead>
                  <tbody>
                    {drivers.map((driver) => (
                      <tr key={driver.id}>
                        <td>
                          <p className="font-medium text-slate-800">
                            {driver.first_name} {driver.last_name}
                          </p>
                          {driver.email && (
                            <p className="text-xs text-slate-500">{driver.email}</p>
                          )}
                        </td>
                        <td className="font-mono">{driver.username}</td>
                        <td>{driver.phone_number}</td>
                        <td>
                          <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            driver.is_active 
                              ? 'bg-green-100 text-green-700' 
                              : 'bg-red-100 text-red-700'
                          }`}>
                            {driver.is_active ? 'Active' : 'Inactive'}
                          </span>
                        </td>
                        <td className="text-sm text-slate-500">
                          {new Date(driver.created_at).toLocaleDateString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </main>

      {/* Create Load Modal */}
      {showCreateModal && (
        <CreateLoadModal
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false);
            fetchLoads();
          }}
        />
      )}

      {/* Assign Driver Modal */}
      {showAssignModal && (
        <AssignDriverModal
          load={showAssignModal}
          drivers={drivers.filter(d => d.is_active)}
          onClose={() => setShowAssignModal(null)}
          onAssigned={() => {
            setShowAssignModal(null);
            fetchLoads();
          }}
        />
      )}

      {/* Meeting Link Modal */}
      {showLinkModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl shadow-2xl max-w-lg w-full mx-4 my-8 animate-fade-in">
            <div className="p-8">
              <div className="flex items-center justify-center relative mb-6">
                <h3 className="text-2xl font-bold text-slate-800 text-center">
                  Meeting Link for {showLinkModal.load.load_number}
                </h3>
                <button
                  onClick={() => setShowLinkModal(null)}
                  className="absolute right-0 text-slate-400 hover:text-slate-600 transition-colors"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              
              <div className="bg-green-50 border border-green-200 rounded-lg p-5 mb-6">
                <div className="flex items-center gap-2 text-green-700 mb-2">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span className="font-medium">Meeting room created!</span>
                </div>
                <p className="text-sm text-green-600">
                  Send the meeting link to the driver via email or copy it to share manually.
                </p>
              </div>

              {showLinkModal.load.driver && (
                <div className="bg-slate-50 rounded-lg p-4 mb-6 space-y-2">
                  <p className="text-sm text-slate-600">
                    <span className="font-medium">Driver:</span> {showLinkModal.load.driver.first_name} {showLinkModal.load.driver.last_name}
                  </p>
                  <p className="text-sm text-slate-600">
                    <span className="font-medium">Phone:</span> {showLinkModal.load.driver.phone_number}
                  </p>
                  {showLinkModal.load.driver.email && (
                    <p className="text-sm text-slate-600">
                      <span className="font-medium">Email:</span> {showLinkModal.load.driver.email}
                    </p>
                  )}
                </div>
              )}
              
              {/* Send Email Button */}
              {showLinkModal.load.driver?.email && (
                <div className="mb-6">
                  <button
                    onClick={handleSendEmail}
                    disabled={sendingEmail || emailSent}
                    className={`w-full px-4 py-3 rounded-lg font-medium flex items-center justify-center gap-2 transition-colors ${
                      emailSent 
                        ? 'bg-green-100 text-green-700 cursor-default'
                        : sendingEmail
                          ? 'bg-blue-100 text-blue-700 cursor-wait'
                          : 'bg-blue-600 text-white hover:bg-blue-700'
                    }`}
                  >
                    {emailSent ? (
                      <>
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        Email Sent Successfully!
                      </>
                    ) : sendingEmail ? (
                      <>
                        <svg className="w-5 h-5 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Sending Email...
                      </>
                    ) : (
                      <>
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                        </svg>
                        Send Link via Email to Driver
                      </>
                    )}
                  </button>
                </div>
              )}

              {!showLinkModal.load.driver?.email && showLinkModal.load.driver && (
                <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-lg">
                  <p className="text-sm text-amber-700">
                    <span className="font-medium">⚠️ No email address:</span> The driver doesn't have an email address on file. Copy the link below to share manually.
                  </p>
                </div>
              )}
              
              <div className="mb-6">
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  Driver Video Call Link:
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={showLinkModal.link}
                    readOnly
                    className="flex-1 px-4 py-2.5 border border-slate-300 rounded-lg bg-slate-50 text-sm font-mono"
                  />
                  <button
                    onClick={() => copyToClipboard(showLinkModal.link)}
                    className="px-4 py-2.5 bg-slate-600 text-white rounded-lg hover:bg-slate-700 transition-colors font-medium flex items-center gap-2"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                    Copy
                  </button>
                </div>
              </div>
              
              <div className="flex gap-3">
                <button 
                  onClick={() => setShowLinkModal(null)} 
                  className="flex-1 px-4 py-2.5 border border-slate-300 text-slate-700 rounded-lg hover:bg-slate-50 transition-colors font-medium"
                >
                  Close
                </button>
                <button 
                  onClick={proceedToMeeting}
                  className="flex-1 px-4 py-2.5 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors font-medium flex items-center justify-center gap-2"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                  Join Meeting
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// Create Load Modal Component
function CreateLoadModal({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const [formData, setFormData] = useState({
    load_number: '',
    description: '',
    pickup_address: '',
    delivery_address: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await adminLoadsApi.create(formData);
      onCreated();
    } catch (err) {
      setError('Failed to create load. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6 animate-fade-in">
        <h2 className="text-xl font-semibold text-slate-800 mb-4">Create New Load</h2>
        
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Load Number <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={formData.load_number}
              onChange={(e) => setFormData({ ...formData, load_number: e.target.value })}
              className="input"
              placeholder="LOAD-001"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="input"
              rows={2}
              placeholder="Load description..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Pickup Address
            </label>
            <input
              type="text"
              value={formData.pickup_address}
              onChange={(e) => setFormData({ ...formData, pickup_address: e.target.value })}
              className="input"
              placeholder="123 Main St, City, State"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Delivery Address
            </label>
            <input
              type="text"
              value={formData.delivery_address}
              onChange={(e) => setFormData({ ...formData, delivery_address: e.target.value })}
              className="input"
              placeholder="456 Oak Ave, City, State"
            />
          </div>

          <div className="flex gap-3 pt-2">
            <button type="button" onClick={onClose} className="btn btn-outline flex-1">
              Cancel
            </button>
            <button type="submit" disabled={loading} className="btn btn-primary flex-1">
              {loading ? 'Creating...' : 'Create Load'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// Assign Driver Modal Component
function AssignDriverModal({ 
  load, 
  drivers, 
  onClose, 
  onAssigned 
}: { 
  load: Load; 
  drivers: Driver[]; 
  onClose: () => void; 
  onAssigned: () => void;
}) {
  const [selectedDriver, setSelectedDriver] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleAssign = async () => {
    if (!selectedDriver) {
      setError('Please select a driver');
      return;
    }

    setError('');
    setLoading(true);

    try {
      await adminLoadsApi.assignDriver(load.id, selectedDriver);
      onAssigned();
    } catch (err) {
      setError('Failed to assign driver. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6 animate-fade-in">
        <h2 className="text-xl font-semibold text-slate-800 mb-2">Assign Driver</h2>
        <p className="text-sm text-slate-500 mb-4">
          Select a driver to assign to load <span className="font-mono font-medium">{load.load_number}</span>
        </p>
        
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {error}
          </div>
        )}

        {drivers.length === 0 ? (
          <p className="text-slate-500 text-center py-4">No active drivers available.</p>
        ) : (
          <div className="space-y-2 max-h-60 overflow-y-auto mb-4">
            {drivers.map((driver) => (
              <label
                key={driver.id}
                className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
                  selectedDriver === driver.id
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-slate-200 hover:border-slate-300'
                }`}
              >
                <input
                  type="radio"
                  name="driver"
                  value={driver.id}
                  checked={selectedDriver === driver.id}
                  onChange={() => setSelectedDriver(driver.id)}
                  className="w-4 h-4 text-blue-600"
                />
                <div className="flex-1">
                  <p className="font-medium text-slate-800">
                    {driver.first_name} {driver.last_name}
                  </p>
                  <p className="text-xs text-slate-500">{driver.phone_number}</p>
                </div>
              </label>
            ))}
          </div>
        )}

        <div className="flex gap-3">
          <button onClick={onClose} className="btn btn-outline flex-1">
            Cancel
          </button>
          <button 
            onClick={handleAssign} 
            disabled={loading || !selectedDriver}
            className="btn btn-primary flex-1"
          >
            {loading ? 'Assigning...' : 'Assign Driver'}
          </button>
        </div>
      </div>
    </div>
  );
}
