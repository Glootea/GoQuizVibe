package com.glootea.goquiz.feature.auth.data

import eu.anifantakis.lib.ksafe.KSafe
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

open class AuthTokenStore(private val ksafe: KSafe? = null) {
    private val tokenState = MutableStateFlow(ksafe?.getDirect(KEY_TOKEN, "") ?: "")
    private val userIdState = MutableStateFlow(ksafe?.getDirect(KEY_USER_ID, "") ?: "")
    private val userRoleState = MutableStateFlow(ksafe?.getDirect(KEY_USER_ROLE, "") ?: "")

    fun observeAccessToken(): Flow<String> = tokenState.asStateFlow()

    open suspend fun accessToken(): String? =
        tokenState.value.takeIf { it.isNotEmpty() }

    open suspend fun put(token: String, userId: String, role: String) {
        ksafe?.let {
            runCatching { it.put(KEY_TOKEN, token) }
            runCatching { it.put(KEY_USER_ID, userId) }
            runCatching { it.put(KEY_USER_ROLE, role) }
        }
        tokenState.value = token
        userIdState.value = userId
        userRoleState.value = role
    }

    open suspend fun clear() {
        ksafe?.let {
            runCatching { it.delete(KEY_TOKEN) }
            runCatching { it.delete(KEY_USER_ID) }
            runCatching { it.delete(KEY_USER_ROLE) }
        }
        tokenState.value = ""
        userIdState.value = ""
        userRoleState.value = ""
    }

    fun cachedUserId(): String = userIdState.value
    fun cachedRole(): String = userRoleState.value

    companion object {
        const val KEY_TOKEN = "auth.access_token"
        const val KEY_USER_ID = "auth.user_id"
        const val KEY_USER_ROLE = "auth.user_role"
    }
}
