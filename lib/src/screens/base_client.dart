import 'package:http/http.dart' as http;
import 'dart:convert';

const baseUrl ="http://ion-suhalim:9988/api/v1";

class ApiService{
  var client = http.Client();
  Future<void> login({
      required String email,
      required String password,
      required String api}) async{
    final url = Uri.parse((baseUrl+api));

    // Create the request body as per your payload struct
    Map<String, String> body = {
      'email': email,
      'password': password,
    };
    try {
      final response = await http.post(
        url,
        headers: {
          'Content-Type': 'application/json',
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        // Handle successful response
        print('User registered successfully');
      } else {
        // Handle errors, e.g. 400, 500, etc.
        print('Failed to register user: ${response.body}');
      }
    } catch (e) {
      print('Error occurred: $e');
    }
  }

  Future<void> registerUser({
    required String name,
    required String email,
    required String password,
    required String phoneNumber,
    required api
  }) async {
    final url = Uri.parse((baseUrl+api));

    // Create the request body as per your payload struct
    Map<String, String> body = {
      'name': name,
      'email': email,
      'password': password,
      'phoneNumber': phoneNumber,
    };
    try {
      final response = await http.post(
        url,
        headers: {
          'Content-Type': 'application/json',
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        // Handle successful response
        print('User registered successfully');
      } else {
        // Handle errors, e.g. 400, 500, etc.
        print('Failed to register user: ${response.body}');
      }
    } catch (e) {
      print('Error occurred: $e');
    }
  }
}

  Future<dynamic> put(String api) async{

  }

  Future<dynamic> delete(String api) async{

  }
