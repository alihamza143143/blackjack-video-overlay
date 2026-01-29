import subprocess
import json

# Number of cards to deal (e.g., 4 for 2 players, 2 cards each)
num_cards = 4

# Prepare the request for the required number of random numbers
request = json.dumps({"nums": num_cards, "gamecode": "blackjack_test"})

cmd = [
    f"{subprocess.os.environ['USERPROFILE']}\\go\\bin\\grpcurl.exe",
    "-plaintext",
    "-import-path", "proto",
    "-proto", "proto/rng.proto",
    "-d", request,
    "localhost:6000",
    "sgc7pb.Rng/getRngs"
]

result = subprocess.run(cmd, capture_output=True, text=True)

try:
    response = json.loads(result.stdout)
    for i, card_num in enumerate(response["rngs"], 1):
        print(f"Card {i}: Random number assigned (overlay): {card_num}")
except Exception:
    print("Raw output:", result.stdout)
    print("Error:", result.stderr)
