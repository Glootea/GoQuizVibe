package com.glootea.goquiz.core.di

import dev.zacsweers.metro.Inject
import dev.zacsweers.metro.SingleIn
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob

@SingleIn(AppScope::class)
@Inject
class AppCoroutineScope(
    dispatcher: CoroutineDispatcher = Dispatchers.Default
) {
    val value: CoroutineScope = CoroutineScope(SupervisorJob() + dispatcher)
}
