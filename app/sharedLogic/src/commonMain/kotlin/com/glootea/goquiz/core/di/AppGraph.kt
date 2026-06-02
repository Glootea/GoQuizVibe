package com.glootea.goquiz.core.di

import com.glootea.goquiz.feature.auth.api.AuthApi
import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import com.glootea.goquiz.feature.auth.domain.usecase.LoginUseCase
import com.glootea.goquiz.feature.auth.domain.usecase.LogoutUseCase
import com.glootea.goquiz.feature.auth.domain.usecase.ObserveMeUseCase
import com.glootea.goquiz.feature.auth.domain.usecase.RegisterUseCase
import com.glootea.goquiz.feature.auth.presentation.AuthStateHolder
import com.glootea.goquiz.feature.auth.presentation.HomeViewModel
import com.glootea.goquiz.feature.auth.presentation.LoginViewModel
import com.glootea.goquiz.feature.auth.presentation.RegisterViewModel
import dev.zacsweers.metro.DependencyGraph

abstract class AppScope private constructor()

@DependencyGraph(AppScope::class)
interface AppGraph {
    val authApi: AuthApi
    val authRepository: AuthRepository
    val loginUseCase: LoginUseCase
    val registerUseCase: RegisterUseCase
    val logoutUseCase: LogoutUseCase
    val observeMeUseCase: ObserveMeUseCase
    val authStateHolder: AuthStateHolder
    val loginViewModel: LoginViewModel
    val registerViewModel: RegisterViewModel
    val homeViewModel: HomeViewModel
}
