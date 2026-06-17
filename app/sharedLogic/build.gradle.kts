import com.codingfeline.buildkonfig.compiler.FieldSpec.Type.STRING
import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    alias(libs.plugins.kotlinMultiplatform)
    alias(libs.plugins.androidMultiplatformLibrary)
    alias(libs.plugins.kotlinSerialization)
    alias(libs.plugins.metro)
    alias(libs.plugins.wire)
    alias(libs.plugins.buildkonfig)
}

kotlin {
    listOf(
        iosArm64(),
        iosSimulatorArm64()
    ).forEach { iosTarget ->
        iosTarget.binaries.framework {
            baseName = "SharedLogic"
            isStatic = true
        }
    }
    jvm()



    android {
       namespace = "com.glootea.goquiz.sharedLogic"
       compileSdk = libs.versions.android.compileSdk.get().toInt()
       minSdk = libs.versions.android.minSdk.get().toInt()

       compilerOptions {
           jvmTarget = JvmTarget.JVM_11
       }
       androidResources {
           enable = true
       }
       withHostTest {
           isIncludeAndroidResources = true
       }
    }

    sourceSets {
        val localMain by creating {
            dependsOn(commonMain.get())
        }
        val onlineMain by creating {
            dependsOn(commonMain.get())
        }
        val localTest by creating {
            dependsOn(commonTest.get())
        }
        val onlineTest by creating {
            dependsOn(commonTest.get())
        }

        commonMain.dependencies {
            implementation(libs.kotlinx.coroutines.core)
            implementation(libs.kotlinx.serialization.json)
            implementation(libs.ksafe)
            implementation(libs.wire.runtime)
            implementation(libs.wire.grpc.client)
            implementation(libs.ktor.http)
            implementation(libs.androidx.lifecycle.viewmodel)
            implementation(libs.androidx.lifecycle.viewmodel.savedstate)
        }
        commonTest.dependencies {
            implementation(libs.kotlin.test)
            implementation(libs.kotlin.reflect)
            implementation(libs.kotlinx.coroutines.test)
        }
    }

    compilerOptions {
        freeCompilerArgs.addAll(listOf("-Xexpect-actual-classes", "-Xexplicit-backing-fields"))
    }
}

val buildkonfigFlavor: String =
    (project.findProperty("buildkonfig.flavor") as String?)?.lowercase() ?: "local"

val targetMain = listOf("androidMain", "jvmMain")
val targetTest = listOf("androidHostTest", "jvmTest")

val sourceMain = when (buildkonfigFlavor) {
    "local" -> "localMain"
    "online" -> "onlineMain"
    else -> error("Unknown buildkonfig.flavor=$buildkonfigFlavor. Use 'local' or 'online'.")
}
val sourceTest = when (buildkonfigFlavor) {
    "local" -> "localTest"
    "online" -> "onlineTest"
    else -> error("Unknown buildkonfig.flavor=$buildkonfigFlavor. Use 'local' or 'online'.")
}

private fun generateDependency(target: String, source: String) = kotlin.sourceSets.getByName(target).dependsOn(kotlin.sourceSets.getByName(source))

targetMain.forEach {generateDependency(it, sourceMain)}

targetTest.forEach {generateDependency(it, sourceTest)}


buildkonfig {
    packageName = "com.glootea.goquiz"
    defaultConfigs {
        buildConfigField(STRING, "FLAVOR", "online", const = true)
    }
    defaultConfigs("local") {
        buildConfigField(STRING, "FLAVOR", "local", const = true)
    }
    defaultConfigs("online") {
        buildConfigField(STRING, "FLAVOR", "online", const = true)
    }
}

wire {
    sourcePath {
        srcDir("src/commonMain/proto")
    }
    kotlin {
        rpcRole = "client"
        rpcCallStyle = "suspending"
    }
}