import { createStore } from 'vuex'
import api from '@/api'

export default createStore({
  state: {
    loading: false,
    contacts: [],
    chatrooms: [],
    sessions: [],
    chatLogs: [],
    currentPage: 1,
    pageSize: 20,
    total: 0,
    apiBase: 'http://127.0.0.1:5030'
  },
  mutations: {
    SET_LOADING(state, loading) {
      state.loading = loading
    },
    SET_CONTACTS(state, contacts) {
      state.contacts = contacts
    },
    SET_CHATROOMS(state, chatrooms) {
      state.chatrooms = chatrooms
    },
    SET_SESSIONS(state, sessions) {
      state.sessions = sessions
    },
    SET_CHAT_LOGS(state, logs) {
      state.chatLogs = logs
    },
    SET_PAGINATION(state, { page, pageSize, total }) {
      state.currentPage = page
      state.pageSize = pageSize
      state.total = total
    },
    SET_API_BASE(state, base) {
      state.apiBase = base
    }
  },
  actions: {
    async fetchContacts({ commit }) {
      commit('SET_LOADING', true)
      try {
        const response = await api.getContacts()
        commit('SET_CONTACTS', response.data)
      } catch (error) {
        console.error('获取联系人失败:', error)
      } finally {
        commit('SET_LOADING', false)
      }
    },
    async fetchChatrooms({ commit }) {
      commit('SET_LOADING', true)
      try {
        const response = await api.getChatrooms()
        commit('SET_CHATROOMS', response.data)
      } catch (error) {
        console.error('获取群聊失败:', error)
      } finally {
        commit('SET_LOADING', false)
      }
    },
    async fetchSessions({ commit }) {
      commit('SET_LOADING', true)
      try {
        const response = await api.getSessions()
        commit('SET_SESSIONS', response.data)
      } catch (error) {
        console.error('获取会话失败:', error)
      } finally {
        commit('SET_LOADING', false)
      }
    },
    async fetchChatLogs({ commit, state }, params) {
      commit('SET_LOADING', true)
      try {
        const response = await api.getChatLogs({
          ...params,
          limit: state.pageSize,
          offset: (state.currentPage - 1) * state.pageSize
        })
        commit('SET_CHAT_LOGS', response.data)
        commit('SET_PAGINATION', {
          page: state.currentPage,
          pageSize: state.pageSize,
          total: response.total || 0
        })
      } catch (error) {
        console.error('获取聊天记录失败:', error)
      } finally {
        commit('SET_LOADING', false)
      }
    }
  },
  getters: {
    isLoading: state => state.loading,
    getContacts: state => state.contacts,
    getChatrooms: state => state.chatrooms,
    getSessions: state => state.sessions,
    getChatLogs: state => state.chatLogs,
    getPagination: state => ({
      current: state.currentPage,
      pageSize: state.pageSize,
      total: state.total
    })
  }
}) 