import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { driverLoadsApi } from '../../api/client';
import type { Load } from '../../api/client';

export default function DriverDashboard() {
  const [loads, setLoads] = useState<Load[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<number | null>(null);
  
  const { logout, driver } = useAuth();
  const navigate = useNavigate();

  const fetchLoads = useCallback(async () => {
    try {
      setLoading(true);
      const data = await driverLoadsApi.getMyLoads();
      setLoads(data.loads || []);
    } catch (error) {
      console.error('Failed to fetch loads:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchLoads();
  }, [fetchLoads]);

  const handleLogout = async () => {
    await logout();
    navigate('/driver/login', { replace: true });
  };

  const handleMarkCompleted = async (loadId: number) => {
    try {
      setUpdating(loadId);
      await driverLoadsApi.markCompleted(loadId);
      await fetchLoads(); // Refresh the list
    } catch (error) {
      console.error('Failed to mark load as completed:', error);
      alert('Failed to update load status. Please try again.');
    } finally {
      setUpdating(null);
    }
  };



  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      'Unassigned': 'bg-slate-100 text-slate-700',
      'Assigned': 'bg-blue-100 text-blue-700',
      'Completed': 'bg-green-100 text-green-700',
    };
    return styles[status] || 'bg-slate-100 text-slate-700';
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 sticky top-0 z-10 shadow-sm">
        <div className="w-full px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-emerald-600 to-emerald-800 rounded-xl flex items-center justify-center">
                <svg 
                  className="w-5 h-5 text-white" 
                  fill="none" 
                  stroke="currentColor" 
                  viewBox="0 0 24 24"
                >
                  <path 
                    strokeLinecap="round" 
                    strokeLinejoin="round" 
                    strokeWidth={2} 
                    d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" 
                  />
                </svg>
              </div>
              <div>
                <h1 className="text-lg font-semibold text-slate-800">My Loads</h1>
                <p className="text-xs text-slate-500">SecuredLinQ Driver Portal</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <span className="text-sm text-slate-600">
                Welcome, <span className="font-medium">{driver?.first_name || driver?.username}</span>
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
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
          <div className="card">
            <p className="text-sm text-slate-500 mb-1">Total Loads</p>
            <p className="text-2xl font-bold text-slate-800">{loads.length}</p>
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

        {/* Refresh button */}
        <div className="flex justify-end mb-4">
          <button
            onClick={fetchLoads}
            disabled={loading}
            className="btn btn-outline flex items-center gap-2 text-sm"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Refresh
          </button>
        </div>

        {/* Loads list */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="spinner"></div>
          </div>
        ) : loads.length === 0 ? (
          <div className="bg-white rounded-xl p-12 shadow-sm border border-slate-200 text-center">
            <div className="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-slate-800 mb-1">No Loads Assigned</h3>
            <p className="text-slate-500">You don't have any loads assigned yet.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {loads.map((load) => (
              <div
                key={load.id}
                className="bg-white rounded-xl p-5 shadow-sm border border-slate-200 hover:shadow-md transition-shadow"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="font-semibold text-slate-800 font-mono">
                        #{load.load_number}
                      </h3>
                      <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusBadge(load.status)}`}>
                        {load.status}
                      </span>
                    </div>
                    
                    {load.description && (
                      <p className="text-slate-600 text-sm mb-3">{load.description}</p>
                    )}

                    <div className="grid grid-cols-2 gap-4 text-sm">
                      {load.pickup_address && (
                        <div>
                          <span className="text-slate-400 block text-xs">Pickup</span>
                          <span className="text-slate-600">{load.pickup_address}</span>
                        </div>
                      )}
                      {load.delivery_address && (
                        <div>
                          <span className="text-slate-400 block text-xs">Delivery</span>
                          <span className="text-slate-600">{load.delivery_address}</span>
                        </div>
                      )}
                    </div>

                    {load.scheduled_date && (
                      <p className="text-xs text-slate-400 mt-3">
                        Scheduled: {formatDate(load.scheduled_date)}
                      </p>
                    )}
                  </div>

                  <div className="flex flex-col gap-2 ml-4">
                    {load.status === 'Assigned' && (
                      <button
                        onClick={() => handleMarkCompleted(load.id)}
                        disabled={updating === load.id}
                        className="btn bg-emerald-600 hover:bg-emerald-700 text-white text-sm px-4 py-2"
                      >
                        {updating === load.id ? (
                          <span className="flex items-center gap-1">
                            <div className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                            Updating...
                          </span>
                        ) : (
                          'Mark Completed'
                        )}
                      </button>
                    )}
                    
                    

                    {load.status === 'Completed' && !load.meeting_started && (
                      <span className="text-xs text-slate-400 text-center">
                        Waiting for admin<br />to start meeting
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}

