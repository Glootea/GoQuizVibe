package com.glootea.goquiz.core.di

import dev.zacsweers.metro.createGraph

object AppGraphFactory {
    fun create(): AppGraph = createGraph<AppGraph>()
}
