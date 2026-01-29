import subprocess
import json

# Prepare the request for 1 random number (nums=1)
request = '{"nums": 1, "gamecode": "test"}'

# Build the grpcurl command
cmd = [
    f"{subprocess.os.environ['USERPROFILE']}\\go\\bin\\grpcurl.exe",
    "-plaintext",
    "-import-path", "proto",
    "-proto", "proto/rng.proto",
    "-d", request,
    "localhost:6000",
    "sgc7pb.Rng/getRngs"
]

# Run the command and capture output
result = subprocess.run(cmd, capture_output=True, text=True)

# Print the output (simulating overlay)
try:
    response = json.loads(result.stdout)
    print("Random number from backend (overlay):", response["rngs"][0])
except Exception:
    print("Raw output:", result.stdout)
    print("Error:", result.stderr)
