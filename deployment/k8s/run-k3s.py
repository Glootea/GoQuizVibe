#!/usr/bin/env python3
import argparse
import subprocess
import sys
from pathlib import Path


def get_kubectl_cmd(use_k3s: bool) -> list[str]:
    return ["k3s", "kubectl"] if use_k3s else ["kubectl"]


def run_kubectl(kubectl: list[str], cmd: list[str], check: bool = False) -> subprocess.CompletedProcess:
    result = subprocess.run([*kubectl, *cmd], capture_output=True, text=True)
    if check and result.returncode != 0:
        print(f"Command failed: {' '.join(kubectl)} {' '.join(cmd)}")
        print(result.stderr)
        sys.exit(1)
    return result


def main():
    parser = argparse.ArgumentParser(description="Deploy GoQuizVibe to k3s")
    parser.add_argument("--use-k3s", action="store_true", help="Use k3s kubectl")
    parser.add_argument(
        "--destroy",
        action="store_true",
        help="Delete resources instead of applying",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Output generated manifests without applying",
    )
    parser.add_argument(
        "--debug",
        action="store_true",
        help="Show pod logs and describe on failure",
    )
    args = parser.parse_args()

    script_dir = Path(__file__).parent.resolve()
    kubectl = get_kubectl_cmd(args.use_k3s)

    overlay_path = script_dir / "overlays" / "k3s"
    generated = run_kubectl(kubectl, ["kustomize", str(overlay_path)])

    if generated.returncode != 0:
        print("=== Kustomize build failed ===")
        print(generated.stderr)
        sys.exit(1)

    if args.dry_run:
        print(generated.stdout)
        return

    if args.destroy:
        print("=== Deleting GoQuizVibe ===")
        subprocess.run([*kubectl, "delete", "--validate=false", "-f", "-"],
                       input=generated.stdout, text=True)
        return

    print("=== Deploying GoQuizVibe ===")
    apply_result = subprocess.run([*kubectl, "apply", "--validate=false", "-f", "-"],
                                  input=generated.stdout, text=True)
    if apply_result.returncode != 0:
        print("Apply failed:")
        print(apply_result.stderr)
        sys.exit(1)

    print("=== Waiting for pods ===")
    for app in ["postgres", "redis", "minio", "backend"]:
        result = subprocess.run(
            [
                *kubectl,
                "wait",
                "--for=condition=ready",
                "pod",
                "-l", f"app={app}",
                "-n",
                "goquizvibe",
                "--timeout=120s",
            ],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0 and args.debug:
            print(f"\n=== Debug: {app} pods ===")
            pods_result = run_kubectl(kubectl, ["get", "pod", "-l", f"app={app}", "-n", "goquizvibe"])
            print(pods_result.stdout)

            pods = pods_result.stdout.strip().split("\n")
            if len(pods) > 1:
                for pod_line in pods[1:]:
                    pod_name = pod_line.split()[0] if pod_line else None
                    if pod_name:
                        print(f"\n--- Logs: {pod_name} ---")
                        run_kubectl(kubectl, ["logs", pod_name, "-n", "goquizvibe"])
                        print(f"\n--- Describe: {pod_name} ---")
                        run_kubectl(kubectl, ["describe", "pod", pod_name, "-n", "goquizvibe"])

    print("=== Deployment complete ===")
    run_kubectl(kubectl, ["get", "pod", "-n", "goquizvibe"])


if __name__ == "__main__":
    main()