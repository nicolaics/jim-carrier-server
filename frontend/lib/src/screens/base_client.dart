import 'package:http/http.dart' as http;
import 'dart:convert';
const baseUrl ="http://ion-suhalim:9988/api/v1";

class ApiService{
  var client = http.Client();

  Future<dynamic> login({
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
         // 'Authorization': 'Bearer ' + token,
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        // Handle successful response
        print('User login successfully');
        return 'success';

      } else {
        // Handle errors, e.g. 400, 500, etc.
        print('Failed to register user: ${response.body}');
      }
    } catch (e) {
      print('Error occurred: $e');
    }
  }

  Future<dynamic> registerUser({
    required String name,
    required String email,
    required String password,
    required String phoneNumber,
    required String verification,
    required api
  }) async {
    final url = Uri.parse((baseUrl+api));

    // Create the request body as per your payload struct
    Map<String, String> body = {
      'name': name,
      'email': email,
      'password': password,
      'phoneNumber': phoneNumber,
      'verificationCode': verification,
    };
    try {
      final response = await http.post(
        url,
        headers: {
          'Content-Type': 'application/json',
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 201) {
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

  Future<dynamic> forgotPw({
    required String email,
    required api
  }) async {
    final url = Uri.parse((baseUrl+api));
    // Create the request body as per your payload struct
    Map<String, String> body = {
      'email': email,
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
        return "success";
      } else {
        // Handle errors, e.g. 400, 500, etc.
        print('Failed to get code: ${response.body}');
      }
    } catch (e) {
      print('Error occurred: $e');
    }
  }

  Future<dynamic> otpCode({
    required String email,
    required api
  }) async {
    final url = Uri.parse((baseUrl+api));
    // Create the request body as per your payload struct
    Map<String, String> body = {
      'email': email,
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
        return "success";
      } else {
        // Handle errors, e.g. 400, 500, etc.
        print('Failed to get code: ${response.body}');
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
