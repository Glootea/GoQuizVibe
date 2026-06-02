package com.glootea.goquiz.feature.auth.data

import com.glootea.goquiz.core.di.AppScope
import com.glootea.goquiz.feature.auth.api.AuthApi
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import com.glootea.goquiz.feature.auth.api.dto.UserDto
import com.glootea.goquiz.feature.auth.domain.model.Role
import com.glootea.goquiz.feature.auth.domain.model.User
import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import dev.zacsweers.metro.ContributesBinding
import dev.zacsweers.metro.Inject

@ContributesBinding(AppScope::class)
@Inject
class AuthRepositoryImpl(
    private val api: AuthApi
) : AuthRepository {

    override suspend fun login(email: String, password: String): User =
        api.login(LoginRequestDto(email = email, password = password)).user.toDomain()

    override suspend fun register(name: String, email: String, password: String): User =
        api.register(RegisterRequestDto(name = name, email = email, password = password)).user.toDomain()

    override suspend fun logout() {
        api.logout()
    }

    override suspend fun me(): User? = api.me()?.user?.toDomain()
}

private fun UserDto.toDomain(): User = User(
    id = id,
    name = name,
    email = email,
    role = when (role) {
        "teacher" -> Role.Teacher
        else -> Role.Student
    }
)
