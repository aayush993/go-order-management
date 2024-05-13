
for ((i=1; ; i++)); do
    echo "Sending request $i..."
    
    # Send request using cURL
    curl -X POST -H "Content-Type: application/json" -d '{"customerId": "1", "productId": "1", "quantity": 1}' "http://localhost:3000/orders"
    
    echo "Request $i completed."

    sleep 5
done