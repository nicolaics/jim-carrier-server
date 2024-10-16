import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:http/http.dart' as http;
import 'package:jim/src/constants/image_strings.dart';
import 'package:jim/src/constants/sizes.dart';
import 'package:jim/src/constants/text_strings.dart';
import 'base_client.dart';


class LoginScreen extends StatefulWidget {
  const LoginScreen({Key? key}) : super(key: key);
  @override
  _LoginScreenState createState() => _LoginScreenState();
}
class _LoginScreenState extends State<LoginScreen> {
  final TextEditingController _emailController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  bool _isPasswordVisible = false;

  final ApiService apiService= ApiService();

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;
    return SafeArea(
      child: Scaffold(
        body: SingleChildScrollView(
            child: Container(
                padding: EdgeInsets.all(tDefaultSize),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SizedBox(height: 50,),
                    Text(
                      tLogInMessage,
                      style: GoogleFonts.anton(fontSize: 40),
                    ),
                    Form(
                        child: Container(
                      padding: EdgeInsets.symmetric(vertical: 20.0),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          TextFormField(
                            controller: _emailController,
                            decoration: InputDecoration(
                                prefixIcon: Icon(Icons.person_outline_outlined),
                                labelText: "Email",
                                hintText: "your_email",
                                border: OutlineInputBorder()),
                          ),
                          SizedBox(
                            height: 30,
                          ),
                          TextFormField(
                            controller: _passwordController,
                            obscureText: !_isPasswordVisible, // Step 3: Use the boolean for visibility
                            decoration: InputDecoration(
                              prefixIcon: Icon(Icons.lock),
                              labelText: "Password",
                              border: OutlineInputBorder(),
                              suffixIcon: IconButton(
                                // Step 4: Toggle visibility and change the icon
                                onPressed: () {
                                  setState(() {
                                    _isPasswordVisible = !_isPasswordVisible;
                                  });
                                },
                                icon: Icon(
                                  _isPasswordVisible
                                      ? Icons.visibility // Show icon when password is visible
                                      : Icons.visibility_off, // Hide icon when password is hidden
                                ),
                              ),
                            ),
                          ),
                          SizedBox(
                            height: 10,
                          ),
                          Align(
                              alignment: Alignment.centerRight,
                              child: TextButton(
                                  onPressed: () {}, child: Text(tForgotPw))),
                          SizedBox(
                              width: double.infinity,
                              child: ElevatedButton(
                                  onPressed: () async {
                                    await apiService.login(
                                      email: _emailController.text,
                                      password: _passwordController.text,
                                      api: '/user/login', // Provide your API base URL
                                    );
                                  },
                                  style: OutlinedButton.styleFrom(
                                      shape: RoundedRectangleBorder(),
                                      backgroundColor: Colors.black),
                                  child: Text(tLogin.toUpperCase(),
                                      style: TextStyle(
                                          color: Colors.white, fontSize: 20)))),
                        ],
                      ),
                    )),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.center,
                      children: [
                        const Text("OR"),
                        SizedBox(height: 10,),
                        SizedBox(
                          width: double.infinity,
                          child: OutlinedButton.icon(
                              icon: Image(image: AssetImage(GoogleImg),width: 20,),
                              onPressed: () {},
                              label: Text('Sign In With Google', style: TextStyle(color: Colors.black))),
                        ),
                        SizedBox(height: 10,),
                        TextButton(
                          onPressed: () {},
                          child: Text.rich(
                            TextSpan(
                              text: "Don't have an Account? ",
                              style: TextStyle(color: Colors.black),
                              children: const[
                                TextSpan(
                                  text: "Signup",
                                  style: TextStyle(color: Colors.blue)
                                )
                              ]
                            )
                          ),),
                      ],
                    )
                  ],
                ))),
      ),
    );
  }
}
