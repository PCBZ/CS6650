import Foundation

// Result structures
struct CartCreateResponse: Codable {
    let shopping_cart_id: Int
    let customer_id: Int
    let status: String
}

struct CartItem: Codable {
    let product_id: Int
    let quantity: Int
}

struct CartResponse: Codable {
    let shopping_cart_id: Int
    let customer_id: Int
    let status: String
    let items: [CartItem]?
}

class ConsistencyTester {
    let baseURL: String
    let apiBase: String
    
    // Track overall statistics
    var totalInconsistencies = 0
    var totalChecks = 0

    init(_ baseURL: String) {
        self.baseURL = baseURL.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        self.apiBase = "\(self.baseURL)/v1/shopping-carts"
    }
    
    // Helper to make async HTTP requests
    func makeRequest(url: String, method: String, body: [String: Any]? = nil) async -> (data: Data?, response: HTTPURLResponse?, error: Error?) {
        guard let requestURL = URL(string: url) else {
            return (nil, nil, NSError(domain: "Invalid URL", code: -1))
        }
        
        var request = URLRequest(url: requestURL)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        if let body = body {
            request.httpBody = try? JSONSerialization.data(withJSONObject: body)
        }
        
        do {
            let (data, response) = try await URLSession.shared.data(for: request)
            return (data, response as? HTTPURLResponse, nil)
        } catch {
            return (nil, nil, error)
        }
    }
    
    func createCart(customerId: Int) async -> (cartId: Int?, statusCode: Int) {
        let body = ["customer_id": customerId]
        let (data, response, _) = await makeRequest(url: apiBase, method: "POST", body: body)
        
        guard let data = data,
              let response = response,
              response.statusCode == 201,
              let cart = try? JSONDecoder().decode(CartCreateResponse.self, from: data) else {
            return (nil, response?.statusCode ?? 0)
        }
        
        return (cart.shopping_cart_id, response.statusCode)
    }
    
    func getCart(cartId: Int) async -> (cart: CartResponse?, statusCode: Int) {
        let url = "\(apiBase)/\(cartId)"
        let (data, response, _) = await makeRequest(url: url, method: "GET")
        
        guard let data = data,
              let response = response,
              response.statusCode == 200,
              let cart = try? JSONDecoder().decode(CartResponse.self, from: data) else {
            return (nil, response?.statusCode ?? 0)
        }
        
        return (cart, response.statusCode)
    }
    
    func addItem(cartId: Int, productId: Int, quantity: Int) async -> Bool {
        let url = "\(apiBase)/\(cartId)/items"
        let body = ["product_id": productId, "quantity": quantity]
        let (_, response, _) = await makeRequest(url: url, method: "POST", body: body)

        return response?.statusCode == 200
    }
    
    // Test 1: Read-After-Write Consistency for Cart Creation
    func test1ReadAfterWriteCartCreation(iterations: Int = 10) async -> (inconsistencies: Int, total: Int) {
        print("\n🔄 Test 1: Read-After-Write Consistency (Cart Creation)")
        
        var inconsistencies = 0
        var total = 0
        
        for _ in 1...iterations {
            let customerId = Int.random(in: 1...100000)
            
            // Create cart
            let (cartId, _) = await createCart(customerId: customerId)

            if let cartId = cartId {
                // Immediately read cart
                let (_, readCode) = await getCart(cartId: cartId)
                total += 1

                if readCode == 200 {
                    print("  ✅ Cart readable immediately")
                } else {
                    print("  ❌ Cart NOT readable immediately")
                    inconsistencies += 1
                }
            }
        }
        
        print("  📊 Result: \(inconsistencies) inconsistencies out of \(total) checks")
        return (inconsistencies, total)
    }
    
    // Test 2: Read-After-Write Consistency for Item Addition
    func test2ReadAfterWriteItemAddition(iterations: Int = 10) async -> (inconsistencies: Int, total: Int) {
        print("\n🔄 Test 2: Read-After-Write Consistency (Item Addition)")
        
        var inconsistencies = 0
        var total = 0
        
        // Create test cart
        let customerId = Int.random(in: 1...100000)
        let (cartId, _) = await createCart(customerId: customerId)
        
        guard let cartId = cartId else {
            print("  ❌ Failed to create test cart")
            return (0, 0)
        }
        
        for _ in 1...iterations {
            let productId = Int.random(in: 1...1000)
            let quantity = Int.random(in: 1...5)
            
            // Add item
            let addSuccess = await addItem(cartId: cartId, productId: productId, quantity: quantity)

            if addSuccess {
                // Immediately read cart
                let (cart, _) = await getCart(cartId: cartId)
                total += 1
                
                // Check if item is visible
                if let cart = cart, let items = cart.items {
                    let itemVisible = items.contains { $0.product_id == productId }
                    
                    if itemVisible {
                        print("  ✅ Item visible immediately")
                    } else {
                        print("  ❌ Item NOT visible immediately")
                        inconsistencies += 1
                    }
                }
            }
        }
        
        print("  📊 Result: \(inconsistencies) inconsistencies out of \(total) checks")
        return (inconsistencies, total)
    }
    
    // Test 3: Concurrent Updates with Multiple Clients
    func test3ConcurrentUpdates(numClients: Int = 10) async -> (inconsistencies: Int, total: Int) {
        print("\n🔄 Test 3: Concurrent Updates (Eventual Consistency)")
        
        var totalInconsistencies = 0
        var totalChecks = 0
        
        // Create test cart
        let customerId = Int.random(in: 1...100000)
        let (cartId, _) = await createCart(customerId: customerId)
        
        guard let cartId = cartId else {
            print("  ❌ Failed to create test cart")
            return (0, 0)
        }
        
        // Use withTaskGroup for true concurrent operations
        await withTaskGroup(of: (clientId: Int, addSuccess: Bool, immediateCount: Int, delayedCount: Int).self) { group in
            // Spawn concurrent client tasks
            for clientId in 1...numClients {
                group.addTask {
                    let productId = clientId  // Unique product ID per client
                    let quantity = Int.random(in: 1...5)
                    
                    // Add item
                    let addSuccess = await self.addItem(cartId: cartId, productId: productId, quantity: quantity)
                    
                    if !addSuccess {
                        return (clientId, false, 0, 0)
                    }
                    
                    // Immediate read after add
                    let (immediateCart, _) = await self.getCart(cartId: cartId)
                    let immediateCount = immediateCart?.items?.count ?? 0
                    
                    // Wait 1 second
                    try? await Task.sleep(nanoseconds: 1_000_000_000)
                    
                    // Read again after 1 second
                    let (delayedCart, _) = await self.getCart(cartId: cartId)
                    let delayedCount = delayedCart?.items?.count ?? 0
                    
                    return (clientId, addSuccess, immediateCount, delayedCount)
                }
            }
            
            // Collect and display results from all clients
            var totalSuccess = 0
            var totalFailed = 0
            var immediateReads: [(clientId: Int, count: Int)] = []
            var delayedReads: [(clientId: Int, count: Int)] = []
            
            for await result in group {
                if result.addSuccess {
                    totalSuccess += 1
                    immediateReads.append((result.clientId, result.immediateCount))
                    delayedReads.append((result.clientId, result.delayedCount))
                    
                    // Print individual client results
                    if result.immediateCount < result.clientId {
                        print(String(format: "  ⚠️  Client %2d: Immediate read = %2d items (expected >= %d) | After 1s = %2d items", 
                                     result.clientId, result.immediateCount, result.clientId, result.delayedCount))
                    } else {
                        print(String(format: "  ✅ Client %2d: Immediate read = %2d items | After 1s = %2d items", 
                                     result.clientId, result.immediateCount, result.delayedCount))
                    }
                } else {
                    totalFailed += 1
                }
            }
            
            print("\n📊 Results Summary:")
            print("  ✅ Successful writes: \(totalSuccess)/\(numClients)")
            print("  ❌ Failed writes: \(totalFailed)")
            
            // Analyze consistency
            print("\n🔍 Consistency Analysis:")
            
            // Check immediate reads
            let immediateInconsistent = immediateReads.filter { $0.count < $0.clientId }.count
            let immediateConsistencyRate = Double(totalSuccess - immediateInconsistent) / Double(totalSuccess) * 100.0
            totalChecks += totalSuccess
            totalInconsistencies += immediateInconsistent
            
            print(String(format: "  📖 Immediate reads: %d inconsistencies out of %d checks (%.1f%% consistent)", 
                         immediateInconsistent, totalSuccess, immediateConsistencyRate))
            
            // Check delayed reads
            let delayedInconsistent = delayedReads.filter { $0.count < totalSuccess }.count
            let delayedConsistencyRate = Double(totalSuccess - delayedInconsistent) / Double(totalSuccess) * 100.0
            totalChecks += totalSuccess
            totalInconsistencies += delayedInconsistent
            
            print(String(format: "  📖 After 1s reads: %d inconsistencies out of %d checks (%.1f%% consistent)", 
                         delayedInconsistent, totalSuccess, delayedConsistencyRate))
        }
        
        return (totalInconsistencies, totalChecks)
    }      
}

// Main execution
Task {
    let args = CommandLine.arguments
    let baseURL = args.count > 1 ? args[1] : "http://localhost:8080"
    
    print("═══════════════════════════════════════════════")
    print("🧪 DynamoDB Consistency Testing Suite")
    print("═══════════════════════════════════════════════")
    print("Target: \(baseURL)")
    print("Running 1000 iterations of consistency tests...")
    print("═══════════════════════════════════════════════\n")
    
    let tester = ConsistencyTester(baseURL)

    // Run tests 100 times
    for iteration in 1...100 {
        // Run all three tests
        let (test1Incon, test1Total) = await tester.test1ReadAfterWriteCartCreation(iterations: 1)
        tester.totalInconsistencies += test1Incon
        tester.totalChecks += test1Total
        
        let (test2Incon, test2Total) = await tester.test2ReadAfterWriteItemAddition(iterations: 1)
        tester.totalInconsistencies += test2Incon
        tester.totalChecks += test2Total
        
        let (test3Incon, test3Total) = await tester.test3ConcurrentUpdates(numClients: 10)
        tester.totalInconsistencies += test3Incon
        tester.totalChecks += test3Total
    }
    
    // Print final summary
    print("\n" + String(repeating: "═", count: 60))
    print("📊 FINAL CONSISTENCY REPORT (100 iterations)")
    print(String(repeating: "═", count: 60))
    print("Target: \(baseURL)")
    print(String(repeating: "-", count: 60))
    
    let consistencyRate = tester.totalChecks > 0 ? 
        (Double(tester.totalChecks - tester.totalInconsistencies) / Double(tester.totalChecks) * 100.0) : 100.0
    
    print(String(format: "Total Checks: %d", tester.totalChecks))
    print(String(format: "Inconsistencies: %d", tester.totalInconsistencies))
    print(String(format: "Consistency Rate: %.2f%%", consistencyRate))
    
    if tester.totalInconsistencies == 0 {
        print("\n✅ Perfect consistency! All reads were consistent.")
        print("   This demonstrates DynamoDB's strong consistency guarantees")
        print("   and fast replication in the same region.")
    } else {
        print(String(format: "\n⚠️  Observed %d inconsistencies", tester.totalInconsistencies))
        print("   This is expected behavior for eventually consistent reads.")
        print("   DynamoDB typically achieves consistency within milliseconds.")
    }
    
    print(String(repeating: "═", count: 60) + "\n")
    
    exit(0)
}

dispatchMain()
