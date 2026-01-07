import axios from 'axios';
import type { AxiosError } from 'axios';

// Base API URL - in development, Vite proxy handles /api routes
const API_BASE_URL = import.meta.env.VITE_API_URL || '';

// Create axios instance with credentials (for cookie auth)
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true, // Important for cookie-based auth
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // Redirect to login if unauthorized (except for auth endpoints)
      const isAuthEndpoint = error.config?.url?.includes('/auth/');
      if (!isAuthEndpoint) {
        window.location.href = '/admin/login';
      }
    }
    return Promise.reject(error);
  }
);

// Types
export interface LoginCredentials {
  username: string;
  password: string;
}

export interface LoginResponse {
  success: boolean;
  message: string;
}

export interface DriverRegisterData {
  username: string;
  password: string;
  phone_number: string;
  first_name?: string;
  last_name?: string;
  email?: string;
}

export interface Driver {
  id: number;
  username: string;
  phone_number: string;
  first_name: string;
  last_name: string;
  email?: string | null; // Now properly serialized as string or null from backend
  is_active: boolean;
  created_at: string;
}

export interface Load {
  id: number;
  load_number: string;
  driver_id?: number;
  driver?: Driver;
  status: 'Unassigned' | 'Assigned' | 'Completed';
  description?: string;
  pickup_address?: string;
  delivery_address?: string;
  scheduled_date?: string;
  completed_at?: string;
  meeting_started: boolean;
  created_by_id: number;
  created_at: string;
  updated_at: string;
}

// Legacy Load type for backward compatibility
export interface LegacyLoad {
  id: number;
  guest_rand: string;
  guest_first_name: string;
  guest_last_name: string;
  guest_phone: string;
  guest_email: string;
  trailer_number: string;
  load_number: string;
  user_id: number;
  created_at: string;
}

export interface MeetingRoom {
  id: number;
  load_id: number;
  loadId?: number; // Alias for load_id
  roomId: string; // Backend returns "roomId" (camelCase)
  channelName: string; // Backend returns "channelName" (camelCase)
  meetingLink: string;
  load_number?: string;
  status: string;
  created_at: string;
  lastJoinedAt?: string;
  guest_rand?: string; // For backward compatibility
  // Legacy fields for backward compatibility
  room_id?: string;
  channel_name?: string;
  guest_id?: number;
}

export interface AgoraTokenRequest {
  channelName: string;
  uid: number;
  role: 'publisher' | 'subscriber';
}

export interface AgoraTokenResponse {
  token: string;
  appId: string;
  channelName: string;
  uid: string;
  expirationTime: number;
}

export interface MediaItem {
  key: string;
  url: string;
  lastModified: string;
  size: number;
  type: 'screenshot' | 'video';
}

// Auth API
export const authApi = {
  // Admin login
  login: async (credentials: LoginCredentials): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>('/api/auth/login', credentials);
    return response.data;
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout');
  },

  validateSession: async (): Promise<{ 
    valid: boolean; 
    user_type?: string;
    user_id?: number;
    admin?: { username: string };
    driver?: Driver;
  }> => {
    const response = await apiClient.get('/api/auth/validate');
    return response.data;
  },

  // Driver registration
  driverRegister: async (data: DriverRegisterData): Promise<{ success: boolean; message: string; driver?: Driver }> => {
    const response = await apiClient.post('/api/auth/driver/register', data);
    return response.data;
  },

  // Driver login
  driverLogin: async (credentials: LoginCredentials): Promise<{ success: boolean; message: string; driver?: Driver }> => {
    const response = await apiClient.post('/api/auth/driver/login', credentials);
    return response.data;
  },
};

// Loads API
export const loadsApi = {
  getAll: async (): Promise<Load[]> => {
    const response = await apiClient.get<{ loads: Load[] }>('/api/loads');
    return response.data.loads;
  },

  getByGuestRand: async (guestRand: string): Promise<Load> => {
    const response = await apiClient.get<{ load: Load }>(`/api/loads/guest?guest_rand=${guestRand}`);
    return response.data.load;
  },

  search: async (query: string): Promise<Load[]> => {
    const response = await apiClient.get<{ loads: Load[] }>(`/api/loads/search?q=${encodeURIComponent(query)}`);
    return response.data.loads;
  },

  getByUserId: async (userId: number): Promise<Load[]> => {
    const response = await apiClient.get<{ loads: Load[] }>(`/api/loads/by-user?user_id=${userId}`);
    return response.data.loads;
  },
};

// Meetings API
export const meetingsApi = {
  create: async (guestRand?: string, loadId?: number): Promise<MeetingRoom> => {
    const payload: { guest_rand?: string; load_id?: number } = {};
    if (guestRand) payload.guest_rand = guestRand;
    if (loadId) payload.load_id = loadId;
    
    const response = await apiClient.post<{ room: MeetingRoom }>('/api/meetings', payload);
    return response.data.room;
  },

  getByRoomId: async (roomId: string): Promise<MeetingRoom> => {
    if (!roomId || roomId === 'undefined') {
      throw new Error('Room ID is required');
    }
    const response = await apiClient.get<{ success: boolean; meetingRoom: MeetingRoom }>(`/api/meetings?roomId=${roomId}`);
    return response.data.meetingRoom;
  },

  end: async (roomId: string): Promise<void> => {
    await apiClient.delete('/api/meetings', { data: { roomId: roomId } });
  },
};

// Agora API
export const agoraApi = {
  getToken: async (request: AgoraTokenRequest): Promise<AgoraTokenResponse> => {
    const response = await apiClient.post<AgoraTokenResponse>('/api/agora/token', request);
    return response.data;
  },

  startRecording: async (data: {
    roomId: string;
    channelName: string;
    uid: string;
    token: string;
  }): Promise<{ resourceId: string; sid: string; uid: string; cname: string; recordingId: string }> => {
    const response = await apiClient.post('/api/agora/recording/start', data);
    return response.data;
  },

  stopRecording: async (data: {
    channelName: string;
    resourceId: string;
    sid: string;
    uid: string;
  }): Promise<void> => {
    await apiClient.post('/api/agora/recording/stop', data);
  },

  queryRecording: async (data: {
    resourceId: string;
    sid: string;
  }): Promise<{ status: string; serverResponse?: Record<string, unknown> }> => {
    const response = await apiClient.get('/api/agora/recording/query', { params: data });
    return response.data;
  },
};

// SMS API
export const smsApi = {
  send: async (phoneNumber: string, message: string): Promise<{ success: boolean; messageSid?: string }> => {
    const response = await apiClient.post('/api/sms/send', { phoneNumber, message });
    return response.data;
  },

  sendMeetingLink: async (phoneNumber: string, meetingLink: string, loadNumber: string): Promise<{ success: boolean; messageSid?: string }> => {
    const response = await apiClient.post('/api/sms/send-meeting-link', { phoneNumber, meetingLink, loadNumber });
    return response.data;
  },
};

// Email API
export const emailApi = {
  sendMeetingLink: async (
    driverEmail: string,
    driverName: string,
    meetingLink: string,
    loadNumber: string
  ): Promise<{ success: boolean; message?: string }> => {
    const response = await apiClient.post('/api/email/send-meeting-link', {
      driverEmail,
      driverName,
      meetingLink,
      loadNumber,
    });
    return response.data;
  },
};

// Media API
export const mediaApi = {
  getLoadMedia: async (guestRand: string): Promise<MediaItem[]> => {
    const response = await apiClient.get<{ media: MediaItem[] }>(`/api/media?guest_rand=${guestRand}`);
    return response.data.media;
  },

  saveScreenshot: async (data: {
    imageData: string;
    roomId?: string;
    loadId?: number;
  }): Promise<{ success: boolean; url?: string; id?: number; s3Key?: string }> => {
    // Backend expects 'screenshot' field, and roomId or loadId
    const response = await apiClient.post('/api/media/screenshot', {
      screenshot: data.imageData,
      roomId: data.roomId,
      loadId: data.loadId,
    });
    return response.data;
  },

  getScreenshotsByLoad: async (loadId: number): Promise<{ success: boolean; screenshots: Array<{ id: number; loadId: number; fileName: string; s3Key: string; videoKey?: string; url: string; type?: string; createdAt: string }> }> => {
    const response = await apiClient.get(`/api/media/screenshots?loadId=${loadId}`);
    return response.data;
  },

  getSignedUrl: async (key: string): Promise<{ url: string }> => {
    const response = await apiClient.get<{ url: string }>(`/api/media/signed-url?key=${encodeURIComponent(key)}`);
    return response.data;
  },
};

// Admin API - Driver Management
export const adminDriversApi = {
  getAll: async (page = 1, pageSize = 20): Promise<{ drivers: Driver[]; total: number }> => {
    const response = await apiClient.get(`/api/admin/drivers?page=${page}&page_size=${pageSize}`);
    return response.data;
  },

  getById: async (id: number): Promise<Driver> => {
    const response = await apiClient.get(`/api/admin/drivers/${id}`);
    return response.data.driver;
  },

  deactivate: async (id: number): Promise<void> => {
    await apiClient.post(`/api/admin/drivers/${id}/deactivate`);
  },

  activate: async (id: number): Promise<void> => {
    await apiClient.post(`/api/admin/drivers/${id}/activate`);
  },
};

// Admin API - Load Management
export const adminLoadsApi = {
  create: async (data: { load_number: string; description?: string; pickup_address?: string; delivery_address?: string }): Promise<Load> => {
    const response = await apiClient.post('/api/admin/loads', data);
    return response.data.load;
  },

  getAll: async (page = 1, pageSize = 20): Promise<{ loads: Load[]; total: number }> => {
    const response = await apiClient.get(`/api/admin/loads?page=${page}&page_size=${pageSize}`);
    return response.data;
  },

  getByStatus: async (status: string, page = 1, pageSize = 20): Promise<{ loads: Load[]; total: number }> => {
    const response = await apiClient.get(`/api/admin/loads/by-status?status=${status}&page=${page}&page_size=${pageSize}`);
    return response.data;
  },

  getById: async (id: number): Promise<Load> => {
    const response = await apiClient.get(`/api/admin/loads/${id}`);
    return response.data.load;
  },

  assignDriver: async (loadId: number, driverId: number): Promise<Load> => {
    const response = await apiClient.post(`/api/admin/loads/${loadId}/assign`, { driver_id: driverId });
    return response.data.load;
  },

  startMeeting: async (loadId: number): Promise<Load> => {
    const response = await apiClient.post(`/api/admin/loads/${loadId}/start-meeting`);
    return response.data.load;
  },

  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/admin/loads/${id}`);
  },
};

// Driver API - Load Management
export const driverLoadsApi = {
  getMyLoads: async (page = 1, pageSize = 20): Promise<{ loads: Load[]; total: number }> => {
    const response = await apiClient.get(`/api/driver/loads?page=${page}&page_size=${pageSize}`);
    return response.data;
  },

  getById: async (id: number): Promise<Load> => {
    const response = await apiClient.get(`/api/driver/loads/${id}`);
    return response.data.load;
  },

  markCompleted: async (loadId: number): Promise<Load> => {
    const response = await apiClient.post(`/api/driver/loads/${loadId}/complete`);
    return response.data.load;
  },

  updateStatus: async (loadId: number, status: string): Promise<Load> => {
    const response = await apiClient.put(`/api/driver/loads/${loadId}/status`, { status });
    return response.data.load;
  },
};

export default apiClient;

